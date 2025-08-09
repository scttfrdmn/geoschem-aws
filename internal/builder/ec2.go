package builder

import (
    "context"
    "fmt"
    "time"
    "encoding/base64"
    "sort"

    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/ec2/types"
    "github.com/aws/aws-sdk-go-v2/aws"
    
    "github.com/your-org/geoschem-aws-platform/internal/common"
)

func (b *Builder) launchBuildInstance(ctx context.Context, config *common.BuildConfig, arch string) (string, error) {
    archConfig := config.Architectures[arch]
    
    // Find latest CIQ Rocky Linux 9 AMI based on architecture
    amiID, err := b.findLatestRockyLinuxAMI(ctx, arch, config.AWS.Region)
    if err != nil {
        return "", fmt.Errorf("finding Rocky Linux AMI: %w", err)
    }
    
    userData := b.generateUserData(config)
    
    input := &ec2.RunInstancesInput{
        ImageId:      aws.String(amiID),
        InstanceType: types.InstanceType(archConfig.InstanceType),
        MinCount:     aws.Int32(1),
        MaxCount:     aws.Int32(1),
        KeyName:      aws.String(config.AWS.KeyPair),
        SecurityGroupIds: []string{config.AWS.SecurityGroup},
        SubnetId:     aws.String(config.AWS.SubnetID),
        UserData:     aws.String(base64.StdEncoding.EncodeToString([]byte(userData))),
        TagSpecifications: []types.TagSpecification{
            {
                ResourceType: types.ResourceTypeInstance,
                Tags: []types.Tag{
                    {Key: aws.String("Name"), Value: aws.String("geoschem-builder")},
                    {Key: aws.String("Project"), Value: aws.String("geoschem-aws")},
                },
            },
        },
    }
    
    result, err := b.ec2Client.RunInstances(ctx, input)
    if err != nil {
        return "", fmt.Errorf("launching instance: %w", err)
    }
    
    instanceID := *result.Instances[0].InstanceId
    fmt.Printf("Launched instance: %s (Rocky Linux 9)\n", instanceID)
    return instanceID, nil
}

// findLatestRockyLinuxAMI finds the latest CIQ Rocky Linux 9 AMI for the specified architecture and region
func (b *Builder) findLatestRockyLinuxAMI(ctx context.Context, arch string, region string) (string, error) {
    var namePattern string
    var architecture string
    
    switch arch {
    case "x86_64":
        namePattern = "Rocky-9-EC2-Base-9.*x86_64*"
        architecture = "x86_64"
    case "arm64":
        namePattern = "Rocky-9-EC2-Base-9.*aarch64*"
        architecture = "arm64"
    default:
        return "", fmt.Errorf("unsupported architecture: %s", arch)
    }
    
    // CIQ is the official publisher of Rocky Linux AMIs (account ID: 679593333241)
    input := &ec2.DescribeImagesInput{
        Owners: []string{"679593333241"},
        Filters: []types.Filter{
            {
                Name:   aws.String("name"),
                Values: []string{namePattern},
            },
            {
                Name:   aws.String("architecture"),
                Values: []string{architecture},
            },
            {
                Name:   aws.String("root-device-type"),
                Values: []string{"ebs"},
            },
            {
                Name:   aws.String("virtualization-type"),
                Values: []string{"hvm"},
            },
            {
                Name:   aws.String("state"),
                Values: []string{"available"},
            },
        },
    }
    
    result, err := b.ec2Client.DescribeImages(ctx, input)
    if err != nil {
        return "", fmt.Errorf("describing Rocky Linux AMIs: %w", err)
    }
    
    if len(result.Images) == 0 {
        return "", fmt.Errorf("no Rocky Linux 9 AMIs found for architecture %s in region %s", arch, region)
    }
    
    // Sort by creation date to get the latest
    sort.Slice(result.Images, func(i, j int) bool {
        return *result.Images[i].CreationDate > *result.Images[j].CreationDate
    })
    
    latestAMI := result.Images[0]
    fmt.Printf("Selected Rocky Linux 9 AMI: %s (%s)\n", *latestAMI.ImageId, *latestAMI.Name)
    
    return *latestAMI.ImageId, nil
}

func (b *Builder) generateUserData(config *common.BuildConfig) string {
    return `#!/bin/bash
# Rocky Linux 9 setup script
dnf update -y

# Install Docker
dnf install -y docker git unzip

# Start and enable Docker
systemctl start docker
systemctl enable docker
usermod -a -G docker rocky

# Install AWS CLI v2 for x86_64
if [ "$(uname -m)" = "x86_64" ]; then
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
else
    curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip"
fi
unzip awscliv2.zip
sudo ./aws/install

# Configure ECR login
aws ecr get-login-password --region ` + config.AWS.Region + ` | docker login --username AWS --password-stdin ` + config.ECRRepository + `

echo "Rocky Linux 9 instance setup complete" > /tmp/setup-complete
`
}

func (b *Builder) waitForInstance(ctx context.Context, instanceID string) error {
    fmt.Printf("Waiting for instance %s to be ready...\n", instanceID)
    
    waiter := ec2.NewInstanceRunningWaiter(b.ec2Client)
    return waiter.Wait(ctx, &ec2.DescribeInstancesInput{
        InstanceIds: []string{instanceID},
    }, 5*time.Minute)
}

func (b *Builder) terminateInstance(ctx context.Context, instanceID string) error {
    fmt.Printf("Terminating instance: %s\n", instanceID)
    
    _, err := b.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
        InstanceIds: []string{instanceID},
    })
    return err
}