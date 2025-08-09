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
)

func main() {
	var (
		profile    = flag.String("profile", "aws", "AWS profile to use")
		region     = flag.String("region", "us-west-2", "AWS region")
		arch       = flag.String("arch", "x86_64", "Architecture (x86_64 or arm64)")
		subnetID   = flag.String("subnet", "", "Subnet ID for instance (required)")
		sgID       = flag.String("security-group", "", "Security Group ID (required)")
		skipCleanup = flag.Bool("keep-instance", false, "Keep instance running after test")
	)
	flag.Parse()

	if *subnetID == "" || *sgID == "" {
		log.Fatal("Both -subnet and -security-group are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
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

	// Create SSH builder
	sshBuilder := builder.NewSSHBuilder(cfg)

	// Create build configuration
	buildConfig := &common.BuildConfig{
		AWS: common.AWSConfig{
			Region:        *region,
			Profile:       *profile,
			SubnetID:      *subnetID,
			SecurityGroup: *sgID,
		},
		Architectures: map[string]common.ArchConfig{
			"x86_64": {
				InstanceType: "t3.medium",
			},
			"arm64": {
				InstanceType: "t4g.medium",
			},
		},
	}

	var instanceID string

	// Cleanup function
	cleanup := func() {
		if instanceID != "" && !*skipCleanup {
			fmt.Println("\nCleaning up instance...")
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
		fmt.Println("\nReceived interrupt, cleaning up...")
		cancel()
		cleanup()
		os.Exit(1)
	}()

	fmt.Printf("🚀 Testing SSH connectivity for architecture: %s\n", *arch)
	fmt.Printf("Using subnet: %s, security group: %s\n", *subnetID, *sgID)

	// Step 1: Launch instance and establish SSH
	fmt.Println("\n=== Step 1: Launch Instance and Establish SSH ===")
	err = sshBuilder.BuildWithSSH(ctx, buildConfig, *arch)
	if err != nil {
		log.Fatalf("Failed to build with SSH: %v", err)
	}

	// Step 2: Basic system info
	fmt.Println("\n=== Step 2: Basic System Information ===")
	commands := []struct {
		desc string
		cmd  string
	}{
		{"Operating System", "cat /etc/os-release | head -2"},
		{"Architecture", "uname -m"},
		{"CPU Info", "lscpu | grep 'Model name'"},
		{"Memory Info", "free -h"},
		{"Disk Space", "df -h /"},
		{"Network Info", "ip route get 1.1.1.1"},
	}

	for _, cmdInfo := range commands {
		fmt.Printf("\n--- %s ---\n", cmdInfo.desc)
		output, err := sshBuilder.ExecuteCommand(ctx, cmdInfo.cmd)
		if err != nil {
			log.Printf("Command failed: %v", err)
			continue
		}
		fmt.Print(output)
	}

	// Step 3: Prepare instance (install Docker, etc.)
	fmt.Println("\n=== Step 3: Prepare Build Environment ===")
	err = sshBuilder.PrepareInstance(ctx)
	if err != nil {
		log.Printf("Failed to prepare instance: %v", err)
		cleanup()
		return
	}

	// Step 4: Test Docker
	fmt.Println("\n=== Step 4: Test Docker Installation ===")
	err = sshBuilder.TestDockerConnection(ctx)
	if err != nil {
		log.Printf("Docker test failed: %v", err)
		cleanup()
		return
	}

	// Step 5: Test AWS CLI
	fmt.Println("\n=== Step 5: Test AWS CLI ===")
	output, err := sshBuilder.ExecuteCommand(ctx, "aws --version")
	if err != nil {
		log.Printf("AWS CLI test failed: %v", err)
	} else {
		fmt.Printf("AWS CLI Version: %s\n", output)
	}

	fmt.Println("\n🎉 SSH connectivity test completed successfully!")

	if *skipCleanup {
		fmt.Println("⚠️  Instance kept running as requested. Don't forget to terminate it manually!")
		// Show connection info
		fmt.Printf("\nTo connect to the instance manually:\n")
		fmt.Printf("ssh -i /tmp/geoschem-builder-%s.pem rocky@<instance-ip>\n", *arch)
	} else {
		cleanup()
	}
}