# GeosChem AWS Platform - Getting Started Guide

This guide will walk you through setting up the complete GeosChem AWS platform from scratch.

## Prerequisites

- AWS CLI configured with your "aws" profile
- Go 1.21+ installed
- Terraform installed
- Git installed
- SSH key pair for EC2 instances

## Step 1: Initialize the Project

```bash
# Create project directory
mkdir geoschem-aws-platform
cd geoschem-aws-platform

# Initialize Go module
go mod init github.com/your-org/geoschem-aws-platform

# Copy all files from the project structure artifact
# (Copy the entire directory structure and all files)

# Install Go dependencies
go mod tidy
```

## Step 2: Configure AWS Profile

Make sure your AWS profile is set up correctly:

```bash
# Configure your AWS profile named "aws"
aws configure --profile aws

# Test the profile
aws sts get-caller-identity --profile aws
```

## Step 3: Create SSH Key Pair

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

## Step 4: Update Configuration Files

Edit `config/aws-config.yaml` with your specific settings:

```yaml
aws:
  profile: "aws"
  region: "us-east-1"  # Change to your preferred region
  
build:
  key_pair: "geoschem-builder-key"
  security_group: "sg-xxxxxxxxx"  # Will be created by Terraform
  subnet_id: "subnet-xxxxxxxxx"   # Will be created by Terraform
  
batch:
  compute_environment: "geoschem-compute"
  job_queue: "geoschem-queue"
  job_definition: "geoschem-job"
```

Edit `config/build-matrix.yaml` to set your ECR repository:

```yaml
ecr_repository: "123456789012.dkr.ecr.us-east-1.amazonaws.com/geoschem"
```

## Step 5: Deploy AWS Infrastructure

```bash
cd terraform

# Initialize Terraform
terraform init

# Plan the deployment
terraform plan

# Deploy infrastructure
terraform apply

# Note the outputs (ECR repository URL, security group ID, subnet ID)
```

Update your config files with the Terraform outputs.

## Step 6: Build Container Images

```bash
# Build a single container for testing
go run cmd/builder/main.go \
    --profile aws \
    --arch x86_64 \
    --compiler gcc13 \
    --mpi openmpi

# Build all containers for x86_64
go run cmd/builder/main.go \
    --profile aws \
    --arch x86_64 \
    --build-all

# Build complete matrix (all architectures and combinations)
go run cmd/builder/main.go \
    --profile aws \
    --build-matrix
```

**Note:** Building the complete matrix will take 1-2 hours and cost approximately $10-20 in EC2 charges.

## Step 7: Deploy Web Interface

```bash
# Make deploy script executable
chmod +x scripts/deploy.sh

# Deploy web interface to S3
./scripts/deploy.sh
```

The script will output the website URL where you can access the platform.

## Step 8: Test the Platform

1. Open the website URL from the deploy script
2. Fill out the job submission form:
   - Job Name: `test-run-1`
   - Compiler: `GCC 13`
   - MPI: `OpenMPI`
   - Architecture: `x86_64`
   - vCPUs: `4`
   - Output S3 Path: `s3://your-results-bucket/test/`
   - Simulation Length: `1` hour
3. Click "Submit Job"
4. Monitor the job status in the interface

## Development Workflow

### Building Individual Containers

```bash
# Intel compiler with Intel MPI on x86_64
go run cmd/builder/main.go --profile aws --arch x86_64 --compiler intel2024 --mpi intelmpi

# GCC with OpenMPI on ARM64 (Graviton)
go run cmd/builder/main.go --profile aws --arch arm64 --compiler gcc13 --mpi openmpi
```

### Updating the Web Interface

```bash
# Make changes to files in web/ directory
# Then redeploy
aws s3 sync web/ s3://your-bucket-name/ --profile aws --delete
```

### Adding New Container Combinations

1. Edit `config/build-matrix.yaml` to add new compiler/MPI combinations
2. Run the builder with your new combination
3. The web interface will automatically detect available containers

## Monitoring and Troubleshooting

### Check Build Logs

```bash
# Monitor EC2 instances during builds
aws ec2 describe-instances \
    --filters "Name=tag:Name,Values=geoschem-builder" \
    --profile aws

# Check instance logs
aws logs describe-log-groups --profile aws
```

### Check Batch Jobs

```bash
# List running jobs
aws batch list-jobs --job-queue geoschem-queue --profile aws

# Describe specific job
aws batch describe-jobs --jobs job-id --profile aws
```

### Check ECR Repository

```bash
# List container images
aws ecr list-images --repository-name geoschem --profile aws

# Get repository URI
aws ecr describe-repositories --repository-names geoschem --profile aws
```

## Cost Optimization Tips

1. **Use Spot Instances**: Configure Batch to use Spot instances for 60-70% savings
2. **Right-size Instances**: Use smallest instance type that meets your needs
3. **Clean Up**: Regularly delete old container images and terminate unused resources
4. **ARM64 Instances**: Use Graviton instances when possible for better price/performance

## Cleanup

To tear down the entire platform:

```bash
# Delete S3 bucket contents
aws s3 rm s3://your-bucket-name --recursive --profile aws

# Destroy Terraform infrastructure
cd terraform
terraform destroy

# Delete ECR images
aws ecr batch-delete-image \
    --repository-name geoschem \
    --image-ids imageTag=latest \
    --profile aws
```

## Next Steps

- Set up automated container builds with GitHub Actions
- Add more GeosChem simulation parameters to the web interface
- Implement job result visualization
- Add email notifications for job completion
- Create custom container builds for specific research needs

## Support

For issues with:
- **Container builds**: Check EC2 instance logs and Docker build output
- **Batch jobs**: Monitor CloudWatch logs for the Batch compute environment
- **Web interface**: Check browser developer console for JavaScript errors
- **AWS permissions**: Verify IAM roles and policies in Terraform output

Remember to monitor your AWS costs, especially during the initial container building phase!