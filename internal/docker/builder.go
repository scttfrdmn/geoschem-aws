package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scttfrdmn/geoschem-aws/internal/ssh"
)

type DockerBuilder struct {
	sshClient *ssh.Client
}

type BuildConfig struct {
	SourceRepo    string // Git repository URL
	SourceBranch  string // Git branch/tag
	DockerfileDir string // Directory containing Dockerfile
	ImageName     string // Final image name
	ImageTag      string // Image tag
	Architecture  string // x86_64 or arm64
	BuildArgs     map[string]string // Docker build arguments
}

// NewDockerBuilder creates a new Docker builder
func NewDockerBuilder(sshClient *ssh.Client) *DockerBuilder {
	return &DockerBuilder{
		sshClient: sshClient,
	}
}

// BuildContainer builds a Docker container on the remote instance
func (db *DockerBuilder) BuildContainer(ctx context.Context, config *BuildConfig) error {
	fmt.Printf("🐳 Starting Docker build for %s:%s (%s)\n", config.ImageName, config.ImageTag, config.Architecture)

	// Step 1: Clone the source repository
	fmt.Println("📥 Cloning source repository...")
	err := db.cloneRepository(ctx, config)
	if err != nil {
		return fmt.Errorf("cloning repository: %w", err)
	}

	// Step 2: Prepare build context
	fmt.Println("📋 Preparing build context...")
	buildDir, err := db.prepareBuildContext(ctx, config)
	if err != nil {
		return fmt.Errorf("preparing build context: %w", err)
	}

	// Step 3: Build the Docker image
	fmt.Println("🔨 Building Docker image...")
	err = db.buildDockerImage(ctx, config, buildDir)
	if err != nil {
		return fmt.Errorf("building Docker image: %w", err)
	}

	// Step 4: Tag the image
	fmt.Println("🏷️  Tagging Docker image...")
	err = db.tagImage(ctx, config)
	if err != nil {
		return fmt.Errorf("tagging image: %w", err)
	}

	fmt.Printf("✅ Docker build completed: %s:%s\n", config.ImageName, config.ImageTag)
	return nil
}

// cloneRepository clones the source repository
func (db *DockerBuilder) cloneRepository(ctx context.Context, config *BuildConfig) error {
	// Clean up any existing source directory
	_, err := db.sshClient.ExecuteCommand(ctx, "rm -rf ~/source")
	if err != nil {
		// Ignore error if directory doesn't exist
	}

	// Clone the repository
	cloneCmd := fmt.Sprintf("git clone --depth 1 --branch %s %s ~/source", 
		config.SourceBranch, config.SourceRepo)
	
	output, err := db.sshClient.ExecuteCommand(ctx, cloneCmd)
	if err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, output)
	}

	fmt.Printf("Repository cloned successfully\n")
	return nil
}

// prepareBuildContext prepares the build context directory
func (db *DockerBuilder) prepareBuildContext(ctx context.Context, config *BuildConfig) (string, error) {
	buildDir := filepath.Join("~/source", config.DockerfileDir)
	
	// Verify Dockerfile exists
	checkCmd := fmt.Sprintf("test -f %s/Dockerfile", buildDir)
	_, err := db.sshClient.ExecuteCommand(ctx, checkCmd)
	if err != nil {
		return "", fmt.Errorf("Dockerfile not found in %s", buildDir)
	}

	// Show build context info
	infoCmd := fmt.Sprintf("cd %s && ls -la && echo '=== Dockerfile ===' && head -20 Dockerfile", buildDir)
	output, err := db.sshClient.ExecuteCommand(ctx, infoCmd)
	if err != nil {
		fmt.Printf("Warning: Could not show build context info: %v\n", err)
	} else {
		fmt.Printf("Build context:\n%s\n", output)
	}

	return buildDir, nil
}

// buildDockerImage builds the Docker image
func (db *DockerBuilder) buildDockerImage(ctx context.Context, config *BuildConfig, buildDir string) error {
	// Construct build command (Rocky Linux 9 uses Podman)
	buildCmd := strings.Builder{}
	buildCmd.WriteString(fmt.Sprintf("cd %s && podman build", buildDir))
	
	// Add build arguments (properly escape values with shell-sensitive characters)
	for key, value := range config.BuildArgs {
		buildCmd.WriteString(fmt.Sprintf(" --build-arg %s='%s'", key, strings.ReplaceAll(value, "'", `'"'"'`)))
	}
	
	// Add platform specification for multi-arch builds (Podman may not need this)
	// platformArch := config.Architecture
	// if platformArch == "arm64" {
	// 	platformArch = "arm64"
	// } else {
	// 	platformArch = "amd64"
	// }
	// buildCmd.WriteString(fmt.Sprintf(" --platform linux/%s", platformArch))
	
	// Add image tag and build context
	buildCmd.WriteString(fmt.Sprintf(" -t %s:%s .", config.ImageName, config.ImageTag))
	
	fmt.Printf("Executing build command: %s\n", buildCmd.String())
	
	// Execute build with streaming output
	err := db.sshClient.ExecuteCommandStream(ctx, buildCmd.String(), os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	return nil
}

// tagImage tags the built image with additional tags
func (db *DockerBuilder) tagImage(ctx context.Context, config *BuildConfig) error {
	// Create architecture-specific tag
	archTag := fmt.Sprintf("%s:%s-%s", config.ImageName, config.ImageTag, config.Architecture)
	tagCmd := fmt.Sprintf("podman tag %s:%s %s", config.ImageName, config.ImageTag, archTag)
	
	output, err := db.sshClient.ExecuteCommand(ctx, tagCmd)
	if err != nil {
		return fmt.Errorf("tagging failed: %w, output: %s", err, output)
	}

	// List final images
	listCmd := fmt.Sprintf("podman images | grep %s", config.ImageName)
	output, err = db.sshClient.ExecuteCommand(ctx, listCmd)
	if err != nil {
		fmt.Printf("Warning: Could not list images: %v\n", err)
	} else {
		fmt.Printf("Built images:\n%s\n", output)
	}

	return nil
}

// PushToECR pushes the built image to Amazon ECR
func (db *DockerBuilder) PushToECR(ctx context.Context, config *BuildConfig, ecrRepository string) error {
	fmt.Printf("📤 Pushing image to ECR: %s\n", ecrRepository)

	// Step 1: Login to ECR
	fmt.Println("🔐 Logging in to ECR...")
	err := db.loginToECR(ctx, ecrRepository)
	if err != nil {
		return fmt.Errorf("ECR login failed: %w", err)
	}

	// Step 2: Tag image for ECR
	fmt.Println("🏷️  Tagging image for ECR...")
	ecrImageName := fmt.Sprintf("%s:%s", ecrRepository, config.ImageTag)
	archECRImageName := fmt.Sprintf("%s:%s-%s", ecrRepository, config.ImageTag, config.Architecture)
	
	// Tag main image
	tagCmd := fmt.Sprintf("podman tag %s:%s %s", config.ImageName, config.ImageTag, ecrImageName)
	output, err := db.sshClient.ExecuteCommand(ctx, tagCmd)
	if err != nil {
		return fmt.Errorf("tagging for ECR failed: %w, output: %s", err, output)
	}

	// Tag architecture-specific image
	tagArchCmd := fmt.Sprintf("podman tag %s:%s-%s %s", config.ImageName, config.ImageTag, config.Architecture, archECRImageName)
	output, err = db.sshClient.ExecuteCommand(ctx, tagArchCmd)
	if err != nil {
		return fmt.Errorf("tagging arch-specific image for ECR failed: %w, output: %s", err, output)
	}

	// Step 3: Push images
	fmt.Println("⬆️  Pushing images to ECR...")
	
	// Push main tag
	pushCmd := fmt.Sprintf("podman push %s", ecrImageName)
	err = db.sshClient.ExecuteCommandStream(ctx, pushCmd, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("pushing main image failed: %w", err)
	}

	// Push architecture-specific tag
	pushArchCmd := fmt.Sprintf("podman push %s", archECRImageName)
	err = db.sshClient.ExecuteCommandStream(ctx, pushArchCmd, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("pushing arch-specific image failed: %w", err)
	}

	fmt.Printf("✅ Successfully pushed to ECR:\n")
	fmt.Printf("   - %s\n", ecrImageName)
	fmt.Printf("   - %s\n", archECRImageName)
	
	return nil
}

// loginToECR authenticates with Amazon ECR
func (db *DockerBuilder) loginToECR(ctx context.Context, ecrRepository string) error {
	// Extract region from ECR repository URL
	// Format: <account>.dkr.ecr.<region>.amazonaws.com/<repo>
	parts := strings.Split(ecrRepository, ".")
	if len(parts) < 4 {
		return fmt.Errorf("invalid ECR repository format: %s", ecrRepository)
	}
	region := parts[3]

	// Get ECR login password and login to Podman
	loginCmd := fmt.Sprintf(
		"aws ecr get-login-password --region %s | podman login --username AWS --password-stdin %s",
		region, strings.Split(ecrRepository, "/")[0])
	
	output, err := db.sshClient.ExecuteCommand(ctx, loginCmd)
	if err != nil {
		return fmt.Errorf("ECR login command failed: %w, output: %s", err, output)
	}

	if !strings.Contains(output, "Login Succeeded") {
		return fmt.Errorf("ECR login did not succeed, output: %s", output)
	}

	fmt.Println("ECR login successful")
	return nil
}

// CleanupImages removes built images to save space
func (db *DockerBuilder) CleanupImages(ctx context.Context, config *BuildConfig) error {
	fmt.Println("🧹 Cleaning up Docker images...")
	
	// Remove built images
	images := []string{
		fmt.Sprintf("%s:%s", config.ImageName, config.ImageTag),
		fmt.Sprintf("%s:%s-%s", config.ImageName, config.ImageTag, config.Architecture),
	}
	
	for _, image := range images {
		cleanupCmd := fmt.Sprintf("podman rmi %s || true", image)
		_, err := db.sshClient.ExecuteCommand(ctx, cleanupCmd)
		if err != nil {
			fmt.Printf("Warning: Failed to remove image %s: %v\n", image, err)
		}
	}

	// Clean up build cache
	_, err := db.sshClient.ExecuteCommand(ctx, "podman system prune -f || true")
	if err != nil {
		fmt.Printf("Warning: Failed to prune build cache: %v\n", err)
	}

	fmt.Println("Cleanup completed")
	return nil
}

// GetImageInfo returns information about built images
func (db *DockerBuilder) GetImageInfo(ctx context.Context, config *BuildConfig) (string, error) {
	// Get image information
	infoCmd := fmt.Sprintf("podman images --format 'table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedSince}}' | grep %s || echo 'No images found'", config.ImageName)
	
	output, err := db.sshClient.ExecuteCommand(ctx, infoCmd)
	if err != nil {
		return "", fmt.Errorf("getting image info: %w", err)
	}

	return output, nil
}