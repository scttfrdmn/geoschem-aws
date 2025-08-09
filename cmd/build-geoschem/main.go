package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/scttfrdmn/geoschem-aws/internal/builder"
	"github.com/scttfrdmn/geoschem-aws/internal/common"
	"github.com/scttfrdmn/geoschem-aws/internal/docker"
	"github.com/scttfrdmn/geoschem-aws/internal/geoschem"
)

func main() {
	var (
		profile       = flag.String("profile", "aws", "AWS profile to use")
		region        = flag.String("region", "us-west-2", "AWS region")
		buildConfig   = flag.String("config", "geoschem-gcc-x86_64", "Build configuration name")
		sourceRepo    = flag.String("repo", "https://github.com/geoschem/GeosChem.git", "Source repository URL")
		sourceBranch  = flag.String("branch", "main", "Source branch/tag")
		imageTag      = flag.String("tag", "latest", "Docker image tag")
		subnetID      = flag.String("subnet", "", "Subnet ID for instance (required)")
		sgID          = flag.String("security-group", "", "Security Group ID (required)")
		ecrRepository = flag.String("ecr", "", "ECR repository URL for pushing (optional)")
		skipBuild     = flag.Bool("skip-build", false, "Skip Docker build (test SSH only)")
		skipPush      = flag.Bool("skip-push", false, "Skip ECR push")
		skipCleanup   = flag.Bool("keep-instance", false, "Keep instance running after build")
		listConfigs   = flag.Bool("list", false, "List available build configurations")
	)
	flag.Parse()

	// List available configurations if requested
	if *listConfigs {
		fmt.Print(geoschem.ListAvailableConfigs())
		return
	}

	// Validate required parameters
	if *subnetID == "" || *sgID == "" {
		log.Fatal("Both -subnet and -security-group are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour) // Extended timeout for builds
	defer cancel()

	// Handle interrupts gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(*profile),
		config.WithRegion(*region),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Get GeosChem build configuration
	geosBuildConfig, err := geoschem.GetBuildConfigByName(*buildConfig)
	if err != nil {
		log.Fatalf("Invalid build configuration: %v", err)
	}

	// Validate configuration
	err = geosBuildConfig.Validate()
	if err != nil {
		log.Fatalf("Build configuration validation failed: %v", err)
	}

	// Create SSH builder
	sshBuilder := builder.NewSSHBuilder(cfg)

	// Create build configuration for AWS
	awsBuildConfig := &common.BuildConfig{
		AWS: common.AWSConfig{
			Region:        *region,
			Profile:       *profile,
			SubnetID:      *subnetID,
			SecurityGroup: *sgID,
		},
		Architectures: map[string]common.ArchConfig{
			"x86_64": {
				InstanceType: "c5.2xlarge", // 8 vCPU for faster builds
			},
			"arm64": {
				InstanceType: "c6g.2xlarge", // 8 vCPU Graviton
			},
		},
	}

	var instanceID string

	// Cleanup function
	cleanup := func() {
		if instanceID != "" && !*skipCleanup {
			fmt.Println("\n🧹 Cleaning up instance...")
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cleanupCancel()
			
			if err := sshBuilder.CleanupInstance(cleanupCtx, instanceID); err != nil {
				log.Printf("Error cleaning up instance: %v", err)
			}
		}
	}

	// Handle interrupts
	go func() {
		<-sigChan
		fmt.Println("\n⚠️  Received interrupt, cleaning up...")
		cancel()
		cleanup()
		os.Exit(1)
	}()

	fmt.Printf("🚀 Starting GeosChem build: %s\n", geosBuildConfig.Name)
	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("   Architecture: %s\n", geosBuildConfig.Architecture)
	fmt.Printf("   Compiler: %s\n", geosBuildConfig.Compiler)
	fmt.Printf("   Source: %s@%s\n", *sourceRepo, *sourceBranch)
	fmt.Printf("   Tag: %s\n", *imageTag)

	// Step 1: Launch instance and establish SSH
	fmt.Println("\n=== Step 1: Launch Build Instance ===")
	instanceID, err = sshBuilder.BuildWithSSH(ctx, awsBuildConfig, geosBuildConfig.Architecture)
	if err != nil {
		log.Fatalf("Failed to setup build instance: %v", err)
	}

	// Step 2: Prepare instance 
	fmt.Println("\n=== Step 2: Prepare Build Environment ===")
	err = sshBuilder.PrepareInstance(ctx)
	if err != nil {
		log.Fatalf("Failed to prepare instance: %v", err)
	}

	// Step 3: Test Docker
	fmt.Println("\n=== Step 3: Verify Docker Installation ===")
	err = sshBuilder.TestDockerConnection(ctx)
	if err != nil {
		log.Fatalf("Docker verification failed: %v", err)
	}

	if !*skipBuild {
		// Step 4: Build Docker container
		fmt.Println("\n=== Step 4: Build GeosChem Container ===")
		
		// Create Docker builder
		dockerBuilder := docker.NewDockerBuilder(sshBuilder.GetSSHClient())
		
		// Convert to Docker build config
		dockerBuildConfig := geosBuildConfig.ToDockerBuildConfig(*sourceRepo, *sourceBranch, *imageTag)
		
		// Execute Docker build
		err = dockerBuilder.BuildContainer(ctx, dockerBuildConfig)
		if err != nil {
			log.Fatalf("Docker build failed: %v", err)
		}

		// Show image information
		imageInfo, err := dockerBuilder.GetImageInfo(ctx, dockerBuildConfig)
		if err != nil {
			log.Printf("Warning: Could not get image info: %v", err)
		} else {
			fmt.Printf("\n📊 Built Images:\n%s\n", imageInfo)
		}

		// Step 5: Push to ECR if requested
		if *ecrRepository != "" && !*skipPush {
			fmt.Println("\n=== Step 5: Push to ECR ===")
			err = dockerBuilder.PushToECR(ctx, dockerBuildConfig, *ecrRepository)
			if err != nil {
				log.Fatalf("ECR push failed: %v", err)
			}
		}

		// Step 6: Cleanup images to save space
		fmt.Println("\n=== Step 6: Cleanup Build Artifacts ===")
		err = dockerBuilder.CleanupImages(ctx, dockerBuildConfig)
		if err != nil {
			log.Printf("Warning: Cleanup failed: %v", err)
		}
	}

	fmt.Println("\n🎉 GeosChem build completed successfully!")
	
	if *skipCleanup {
		fmt.Println("⚠️  Instance kept running as requested.")
		fmt.Printf("💡 To connect: ssh -i /tmp/geoschem-builder-%s.pem rocky@<instance-ip>\n", geosBuildConfig.Architecture)
		fmt.Println("🗑️  Don't forget to terminate the instance manually!")
	} else {
		cleanup()
	}
}

// Add method to SSH builder to expose SSH client
func init() {
	// This is a workaround to expose the SSH client from SSH builder
	// In a real implementation, we'd modify the SSH builder struct
}