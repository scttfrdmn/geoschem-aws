package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/scttfrdmn/geoschem-aws/internal/common"
	"github.com/scttfrdmn/geoschem-aws/internal/ssh"
)

type SSHBuilder struct {
	*Builder
	keyPairManager *ssh.KeyPairManager
	sshClient      *ssh.Client
	instanceID     string
}

// NewSSHBuilder creates a new SSH-enabled builder
func NewSSHBuilder(cfg aws.Config) *SSHBuilder {
	builder := NewFromConfig(cfg, cfg.Region)
	return &SSHBuilder{
		Builder:        builder,
		keyPairManager: ssh.NewKeyPairManager(builder.ec2Client),
	}
}

// BuildWithSSH launches an instance and establishes SSH connection for building
func (sb *SSHBuilder) BuildWithSSH(ctx context.Context, config *common.BuildConfig, arch string) (string, error) {
	// Setup key pair for SSH access
	keyPairName := fmt.Sprintf("geoschem-builder-%s", arch)
	privateKeyPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.pem", keyPairName))

	// Ensure key pair exists
	err := sb.keyPairManager.GetOrCreateKeyPair(ctx, keyPairName, privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("setting up key pair: %w", err)
	}

	// Update config to use our key pair
	config.AWS.KeyPair = keyPairName

	// Launch the build instance
	instanceID, err := sb.launchBuildInstance(ctx, config, arch)
	if err != nil {
		return "", fmt.Errorf("launching build instance: %w", err)
	}

	sb.instanceID = instanceID // Store for later use
	fmt.Printf("Launched build instance: %s\n", instanceID)

	// Wait for instance to be running and get public IP
	publicIP, err := sb.waitForInstanceReady(ctx, instanceID)
	if err != nil {
		return instanceID, fmt.Errorf("waiting for instance: %w", err)
	}

	fmt.Printf("Instance ready with public IP: %s\n", publicIP)

	// Setup SSH client
	sb.sshClient, err = ssh.NewClient(publicIP, "rocky", privateKeyPath)
	if err != nil {
		return instanceID, fmt.Errorf("creating SSH client: %w", err)
	}

	// Wait for SSH to be available (instance needs to boot)
	fmt.Println("Waiting for SSH connection...")
	err = sb.sshClient.WaitForConnection(ctx, publicIP, 30) // 30 retries = ~5 minutes
	if err != nil {
		return instanceID, fmt.Errorf("establishing SSH connection: %w", err)
	}

	fmt.Println("SSH connection established!")

	// Test SSH connection
	err = sb.sshClient.TestConnection(ctx)
	if err != nil {
		return instanceID, fmt.Errorf("testing SSH connection: %w", err)
	}

	fmt.Println("SSH connection verified!")
	return instanceID, nil
}

// waitForInstanceReady waits for instance to be running and returns public IP
func (sb *SSHBuilder) waitForInstanceReady(ctx context.Context, instanceID string) (string, error) {
	waiter := ec2.NewInstanceRunningWaiter(sb.ec2Client)
	
	// Wait for instance to be running (max 5 minutes)
	err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, 5*time.Minute)
	if err != nil {
		return "", fmt.Errorf("waiting for instance to be running: %w", err)
	}

	// Get instance details to retrieve public IP
	result, err := sb.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", fmt.Errorf("describing instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found in describe result")
	}

	instance := result.Reservations[0].Instances[0]
	if instance.PublicIpAddress == nil {
		return "", fmt.Errorf("instance has no public IP address")
	}

	return *instance.PublicIpAddress, nil
}

// ExecuteCommand runs a command on the build instance via SSH
func (sb *SSHBuilder) ExecuteCommand(ctx context.Context, command string) (string, error) {
	if sb.sshClient == nil {
		return "", fmt.Errorf("SSH client not initialized")
	}

	return sb.sshClient.ExecuteCommand(ctx, command)
}

// ExecuteCommandStream runs a command and streams output in real-time
func (sb *SSHBuilder) ExecuteCommandStream(ctx context.Context, command string) error {
	if sb.sshClient == nil {
		return fmt.Errorf("SSH client not initialized")
	}

	return sb.sshClient.ExecuteCommandStream(ctx, command, os.Stdout, os.Stderr)
}

// UploadFile uploads a file to the build instance
func (sb *SSHBuilder) UploadFile(ctx context.Context, localPath, remotePath string) error {
	if sb.sshClient == nil {
		return fmt.Errorf("SSH client not initialized")
	}

	return sb.sshClient.UploadFile(ctx, localPath, remotePath)
}

// PrepareInstance sets up the instance for building (install Docker, etc.)
func (sb *SSHBuilder) PrepareInstance(ctx context.Context, skipUpdate bool) error {
	fmt.Println("Preparing build instance...")

	if !skipUpdate {
		// Clean package cache and update system packages with conflict resolution
		fmt.Println("Cleaning package cache and updating system packages...")
		err := sb.ExecuteCommandStream(ctx, "sudo dnf clean all && sudo dnf update -y --allowerasing")
		if err != nil {
			return fmt.Errorf("updating packages: %w", err)
		}

		// Check if kernel was updated and reboot if necessary
		fmt.Println("Checking if reboot is needed...")
		needsReboot, err := sb.ExecuteCommand(ctx, "dnf needs-restarting -r; echo $?")
		if err != nil {
			fmt.Printf("Warning: Could not check reboot status: %v\n", err)
		} else if strings.Contains(needsReboot, "1") {
			fmt.Println("Kernel update detected, rebooting instance...")
			// Initiate reboot
			_, err := sb.ExecuteCommand(ctx, "sudo reboot")
			if err != nil {
				fmt.Printf("Warning: Reboot command failed: %v\n", err)
			}
			
			// Wait for reboot and reconnect
			fmt.Println("Waiting for instance to reboot...")
			time.Sleep(30 * time.Second) // Wait for reboot to begin
			
			// Re-establish SSH connection
			publicIP, err := sb.waitForInstanceReady(ctx, sb.instanceID)
			if err != nil {
				return fmt.Errorf("waiting for instance after reboot: %w", err)
			}
			
			err = sb.sshClient.WaitForConnection(ctx, publicIP, 30)
			if err != nil {
				return fmt.Errorf("reconnecting SSH after reboot: %w", err)
			}
			
			fmt.Println("Successfully reconnected after reboot!")
		}
	} else {
		fmt.Println("Skipping system package update for faster testing...")
	}

	// Install Docker/Podman (Rocky Linux 9 uses Podman with Docker compatibility)
	fmt.Println("Installing container runtime...")
	containerInstall := "sudo dnf install -y podman git unzip && sudo systemctl enable --now podman.socket && sudo usermod -aG wheel rocky"
	err := sb.ExecuteCommandStream(ctx, containerInstall)
	if err != nil {
		return fmt.Errorf("installing container runtime: %w", err)
	}

	// Install AWS CLI 2.x (as requested by user - dnf version is old)
	fmt.Println("Installing AWS CLI 2.x...")
	awsInstall := "curl \"https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip\" -o \"awscliv2.zip\" && unzip awscliv2.zip && sudo ./aws/install && rm -rf aws awscliv2.zip && aws --version"
	err = sb.ExecuteCommandStream(ctx, awsInstall)
	if err != nil {
		return fmt.Errorf("installing AWS CLI: %w", err)
	}

	// Install additional build tools
	fmt.Println("Installing build tools...")
	err = sb.ExecuteCommandStream(ctx, "sudo dnf install -y make gcc gcc-gfortran")
	if err != nil {
		return fmt.Errorf("installing build tools: %w", err)
	}

	fmt.Println("Instance preparation completed!")
	return nil
}

// TestDockerConnection verifies container runtime is working
func (sb *SSHBuilder) TestDockerConnection(ctx context.Context) error {
	fmt.Println("Testing container runtime...")
	
	// Test basic container command (Rocky Linux 9 uses Podman)
	_, err := sb.ExecuteCommand(ctx, "podman --version")
	if err != nil {
		return fmt.Errorf("testing container runtime: %w", err)
	}

	// Enable Docker compatibility alias if not already set
	fmt.Println("Setting up Docker compatibility alias...")
	err = sb.ExecuteCommandStream(ctx, "sudo dnf install -y podman-docker")
	if err != nil {
		fmt.Printf("Warning: Could not install docker alias: %v\n", err)
	}

	// Pull and run a small test image using podman
	fmt.Println("Testing container pull and run...")
	err = sb.ExecuteCommandStream(ctx, "podman run --rm hello-world")
	if err != nil {
		return fmt.Errorf("testing container functionality: %w", err)
	}

	fmt.Println("Container runtime verified!")
	return nil
}

// GetSSHClient returns the SSH client for direct use
func (sb *SSHBuilder) GetSSHClient() *ssh.Client {
	return sb.sshClient
}

// CleanupInstance terminates the build instance
func (sb *SSHBuilder) CleanupInstance(ctx context.Context, instanceID string) error {
	if sb.sshClient != nil {
		sb.sshClient.Close()
	}

	fmt.Printf("Terminating instance: %s\n", instanceID)
	
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err := sb.ec2Client.TerminateInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("terminating instance: %w", err)
	}

	// Wait for termination to complete
	waiter := ec2.NewInstanceTerminatedWaiter(sb.ec2Client)
	err = waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("waiting for instance termination: %w", err)
	}

	fmt.Printf("Instance %s terminated successfully\n", instanceID)
	return nil
}