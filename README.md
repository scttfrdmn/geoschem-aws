# GeosChem AWS Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://semver.org)

A comprehensive platform for running GeosChem atmospheric chemistry simulations on AWS using containerized builds and AWS Batch, now powered by **Rocky Linux 9**.

**Version**: 0.1.0-alpha (following [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html))

> **⚠️ Alpha Release**: This is an early alpha version with core foundations in place but missing critical functionality. See [ROADMAP.md](ROADMAP.md) for development plan.

## Features

### ✅ **Implemented (v0.1.0-alpha)**
- **Rocky Linux 9 Integration**: CIQ official AMI support with dynamic lookup
- **AWS Quota Management**: Proactive quota checking and increase guidance  
- **Intelligent Instance Recommendations**: Cost-optimized instance type selection
- **Comprehensive AWS Setup**: Step-by-step configuration with IAM policies
- **Multi-Architecture Support**: x86_64 and ARM64 (Graviton) containers
- **Enterprise Foundation**: MIT licensed, semantic versioning, full documentation

### 🚧 **In Development (Target: v0.1.0)**
- Container building on native AWS instances (currently stubbed)
- AWS Batch integration for job execution (planned)
- Terraform infrastructure deployment (planned)  
- Web interface for job submission and monitoring (planned)
- Integration with AWS Open Data for input datasets (planned)

## Prerequisites

### AWS Account Requirements
This platform requires an active AWS account with appropriate permissions. You'll need:

- **AWS Account**: Active AWS account with billing enabled
- **AWS CLI**: Installed and configured (v2 recommended)
- **AWS Profile**: Named profile with appropriate IAM permissions
- **Key Pair**: EC2 key pair for build instances
- **VPC/Networking**: Default VPC or custom networking setup

### Required AWS Services
The platform uses the following AWS services (ensure they're available in your region):
- **EC2**: For container build instances
- **ECR**: For container image registry
- **Batch**: For job execution
- **S3**: For results storage and web hosting
- **IAM**: For service permissions
- **VPC**: For networking (default VPC acceptable)

## AWS Setup Guide

### 1. Install AWS CLI
```bash
# macOS (using Homebrew)
brew install awscli

# Linux/Windows - Download from AWS
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip && sudo ./aws/install
```

### 2. Configure AWS Profile
Create a named AWS profile called 'aws' (or customize in config files):

```bash
# Configure the 'aws' profile
aws configure --profile aws

# You'll be prompted for:
# AWS Access Key ID: [Your access key]
# AWS Secret Access Key: [Your secret key] 
# Default region name: [e.g., us-east-1]
# Default output format: [json]
```

### 3. Test AWS Configuration
```bash
# Verify your profile works
aws sts get-caller-identity --profile aws

# Expected output:
# {
#     "UserId": "AIDACKCEVSQ6C2EXAMPLE",
#     "Account": "123456789012", 
#     "Arn": "arn:aws:iam::123456789012:user/your-username"
# }
```

### 4. Required IAM Permissions

The platform requires the following AWS permissions. You can either:
- **Option A**: Use an existing user with `AdministratorAccess` (easiest for testing)
- **Option B**: Create a custom IAM policy with minimal required permissions (production recommended)

#### Option B: Minimal IAM Policy

Create an IAM policy with these permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "EC2Permissions",
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeImages",
                "ec2:DescribeInstances", 
                "ec2:DescribeInstanceStatus",
                "ec2:DescribeKeyPairs",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeSubnets",
                "ec2:DescribeVpcs",
                "ec2:RunInstances",
                "ec2:TerminateInstances",
                "ec2:CreateTags"
            ],
            "Resource": "*"
        },
        {
            "Sid": "ECRPermissions", 
            "Effect": "Allow",
            "Action": [
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:BatchGetImage",
                "ecr:DescribeRepositories",
                "ecr:ListImages",
                "ecr:DescribeImages",
                "ecr:BatchDeleteImage",
                "ecr:InitiateLayerUpload",
                "ecr:UploadLayerPart",
                "ecr:CompleteLayerUpload",
                "ecr:PutImage"
            ],
            "Resource": "*"
        },
        {
            "Sid": "BatchPermissions",
            "Effect": "Allow", 
            "Action": [
                "batch:DescribeComputeEnvironments",
                "batch:DescribeJobDefinitions", 
                "batch:DescribeJobQueues",
                "batch:DescribeJobs",
                "batch:ListJobs",
                "batch:SubmitJob",
                "batch:TerminateJob",
                "batch:CancelJob"
            ],
            "Resource": "*"
        },
        {
            "Sid": "S3Permissions",
            "Effect": "Allow",
            "Action": [
                "s3:CreateBucket",
                "s3:DeleteBucket", 
                "s3:GetBucketLocation",
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject",
                "s3:ListBucket",
                "s3:PutBucketWebsite",
                "s3:PutBucketPolicy"
            ],
            "Resource": [
                "arn:aws:s3:::geoschem-*",
                "arn:aws:s3:::geoschem-*/*"
            ]
        },
        {
            "Sid": "IAMPermissions",
            "Effect": "Allow",
            "Action": [
                "iam:CreateRole",
                "iam:DeleteRole", 
                "iam:GetRole",
                "iam:PassRole",
                "iam:AttachRolePolicy",
                "iam:DetachRolePolicy",
                "iam:CreateInstanceProfile",
                "iam:DeleteInstanceProfile",
                "iam:AddRoleToInstanceProfile",
                "iam:RemoveRoleFromInstanceProfile"
            ],
            "Resource": [
                "arn:aws:iam::*:role/geoschem-*",
                "arn:aws:iam::*:instance-profile/geoschem-*"
            ]
        },
        {
            "Sid": "STSPermissions",
            "Effect": "Allow",
            "Action": [
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        }
    ]
}
```

### 5. Create EC2 Key Pair
```bash
# Create key pair for EC2 instances
aws ec2 create-key-pair \
    --key-name geoschem-builder-key \
    --profile aws \
    --query 'KeyMaterial' \
    --output text > ~/.ssh/geoschem-builder-key.pem

# Set proper permissions
chmod 400 ~/.ssh/geoschem-builder-key.pem
```

## Quick Start (Alpha)

> **Note**: This alpha version provides AWS setup validation and instance recommendations, but doesn't yet build containers or run jobs. See [DEVELOPMENT.md](DEVELOPMENT.md) to contribute to completing the implementation.

Once AWS is configured:

1. **Clone and Setup**
   ```bash
   git clone <repository-url>
   cd geoschem-aws
   ```

2. **Test Current Functionality**
   ```bash
   # Check version and basic functionality
   go run cmd/builder/main.go --version
   
   # Get instance type recommendations
   go run cmd/builder/main.go --recommend-instance \
       --grid-resolution 4x5 \
       --species-count 150 \
       --priority cost \
       --profile aws

   # Check AWS quotas  
   go run cmd/builder/main.go --check-quotas --profile aws --region us-east-1
   ```

3. **Configure the Platform**
   Edit `config/build-matrix.yaml` to set your AWS profile, region, and ECR repository:
   ```yaml
   aws:
     profile: "aws"
     region: "us-east-1"  # Your preferred region
   ecr_repository: "123456789012.dkr.ecr.us-east-1.amazonaws.com/geoschem"
   ```

3. **Deploy Infrastructure**
   ```bash
   cd terraform
   terraform init
   terraform apply
   ```

4. **Build Containers**
   ```bash
   # Single container
   go run cmd/builder/main.go --profile aws --region us-east-1 --arch x86_64 --compiler gcc13 --mpi openmpi
   
   # All combinations for an architecture
   go run cmd/builder/main.go --profile aws --region us-east-1 --arch x86_64 --build-all
   
   # Complete matrix
   go run cmd/builder/main.go --profile aws --region us-east-1 --build-matrix
   ```

## Container Variants (Rocky Linux 9)

### x86_64 Architecture
- `geoschem:intel2024-intelmpi`
- `geoschem:intel2024-openmpi`
- `geoschem:gcc13-openmpi`
- `geoschem:gcc13-mpich`
- `geoschem:aocc4-openmpi`

### ARM64 Architecture (Graviton)
- `geoschem:gcc13-openmpi-arm64`
- `geoschem:gcc13-mpich-arm64`

## Rocky Linux 9 Benefits

- **Enterprise Stability**: RHEL 9 compatibility for scientific workloads
- **Official CIQ Images**: Uses CIQ's official Rocky Linux 9 AMIs (Account ID: 679593333241)
- **Better HPC Support**: Enhanced scientific computing package ecosystem
- **Security**: Regular enterprise-grade security updates
- **Long-term Support**: Predictable lifecycle for research environments

## Configuration

### AWS Profile and Region
The platform uses configurable AWS profiles and regions. You can:
- Set defaults in `config/build-matrix.yaml`
- Override via command line: `--profile myprofile --region us-west-2`

### Example Build Commands
```bash
# Using default profile 'aws' with custom region
go run cmd/builder/main.go --region us-west-2 --arch x86_64 --compiler gcc13 --mpi openmpi

# Using custom profile
go run cmd/builder/main.go --profile production --region eu-west-1 --build-matrix
```

## Usage

### Building Containers
```bash
# Build single combination
go run cmd/builder/main.go --profile aws --arch x86_64 --compiler gcc13 --mpi openmpi

# Build all for architecture
go run cmd/builder/main.go --profile aws --arch x86_64 --build-all

# Build complete matrix
go run cmd/builder/main.go --profile aws --build-matrix
```

### Running Simulations
Access the web interface at your S3 bucket URL and submit jobs through the simple form.

## Development

### Project Structure
```
geoschem-aws-platform/
├── cmd/builder/           # Main builder command
├── internal/
│   ├── builder/          # Builder logic with Rocky Linux support
│   └── common/           # Shared configuration
├── config/               # Configuration files
├── docker/               # Rocky Linux 9 Dockerfiles
└── terraform/            # Infrastructure as code
```

### Dependencies
- Go 1.21+
- AWS CLI configured with your profile
- Terraform (for infrastructure)

## Cost Optimization

1. **Use CIQ Rocky Linux 9**: Official images with enterprise support
2. **Spot Instances**: Configure Batch to use Spot instances for 60-70% savings
3. **ARM64 Instances**: Use Graviton instances when possible for better price/performance
4. **Right-size**: Use smallest instance type that meets your needs

## Troubleshooting

### Rocky Linux 9 Specific
- Default user is `rocky` (not `ec2-user`)
- Uses `dnf` package manager (not `yum`)
- Containers run as non-root `geoschem` user

### AWS Configuration Issues

#### Profile and Authentication
```bash
# Test your AWS profile
aws sts get-caller-identity --profile aws

# If this fails, reconfigure:
aws configure --profile aws

# Check your credentials file
cat ~/.aws/credentials
cat ~/.aws/config
```

#### Permission Errors
```bash
# Test specific permissions
aws ec2 describe-images --owners=679593333241 --profile aws
aws ecr describe-repositories --profile aws

# If permission denied, ensure your IAM user/role has required policies
```

#### Rocky Linux AMI Issues
```bash
# List available Rocky Linux 9 AMIs in your region
aws ec2 describe-images \
    --owners=679593333241 \
    --filters "Name=name,Values=Rocky-9-EC2-Base-9.*x86_64*" \
    --query 'Images[0].[ImageId,Name,CreationDate]' \
    --profile aws

# If no results, Rocky Linux 9 may not be available in your region
# Try a different region or contact CIQ
```

#### ECR Repository Setup
```bash
# Create ECR repository if it doesn't exist
aws ecr create-repository \
    --repository-name geoschem \
    --profile aws

# Get repository URI for config
aws ecr describe-repositories \
    --repository-names geoschem \
    --query 'repositories[0].repositoryUri' \
    --output text \
    --profile aws
```

### Common Issues
- **AMI not found**: Ensure CIQ Rocky Linux 9 is available in your target region
- **Profile errors**: Verify AWS profile configuration with `aws sts get-caller-identity --profile aws`
- **Region mismatch**: Ensure ECR repository matches your build region
- **Permission denied**: Check IAM policies match the minimal permissions above
- **Key pair missing**: Ensure EC2 key pair exists in your target region

## Support

- Check CloudWatch logs for build details
- Monitor EC2 instances during container builds
- Verify IAM permissions for cross-service access

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Scott Friedman

## Contributing

This project is in active development! We welcome contributions to help complete the implementation.

### 🚀 **High Impact Areas**
- **Terraform Infrastructure**: Create AWS resource deployment ([DEVELOPMENT.md](DEVELOPMENT.md))
- **Docker Build Implementation**: Replace stubbed build process with real SSH-based execution
- **AWS Batch Integration**: Connect job submission to actual Batch API  
- **Instance Type Benchmarking**: Validate recommendations with real GeosChem runs

### 📋 **Getting Started**
1. Read [DEVELOPMENT.md](DEVELOPMENT.md) for detailed development setup
2. Check [ROADMAP.md](ROADMAP.md) for the complete development plan
3. Pick an issue from the development TODO list
4. Join our community discussions about atmospheric chemistry on AWS

### 🧪 **Testing & Validation**
Help us validate the theoretical recommendations with real workloads:
- Benchmark GeosChem performance across instance types
- Test ARM64 (Graviton) compatibility and performance
- Validate cost optimization strategies

## Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html). For the versions available, see the [CHANGELOG.md](CHANGELOG.md) file.