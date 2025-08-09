# AWS Setup Guide for GeosChem Platform

This guide provides detailed instructions for setting up AWS access for the GeosChem AWS Platform.

## Overview

The GeosChem AWS Platform requires access to multiple AWS services and proper IAM permissions. This document walks through the complete setup process.

## Prerequisites

1. **Active AWS Account** with billing enabled
2. **AWS CLI v2** installed on your local machine
3. **Administrative access** to create IAM users/policies (or existing credentials)

## Step-by-Step Setup

### 1. Install AWS CLI v2

**macOS (Homebrew):**
```bash
brew install awscli
```

**Linux:**
```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

**Windows:**
Download and run the AWS CLI MSI installer from the AWS documentation.

### 2. Create IAM User (if needed)

If you don't have programmatic access credentials:

1. Login to AWS Console → IAM → Users → Create User
2. User name: `geoschem-platform-user`
3. Select: "Provide user access to the AWS Management Console" (optional)
4. Select: "I want to create an IAM user"
5. Attach policies (see step 3)
6. Download credentials CSV or save Access Key ID and Secret Access Key

### 3. Create IAM Policy

#### Option A: Use Administrative Access (Easiest)
Attach the `AdministratorAccess` managed policy to your user.

#### Option B: Create Minimal Policy (Production Recommended)

1. Go to IAM → Policies → Create Policy
2. Select JSON tab
3. Copy the contents of [`docs/iam-policy.json`](iam-policy.json)
4. Name: `GeoChemPlatformPolicy`
5. Description: "Minimal permissions for GeosChem AWS Platform"
6. Create Policy
7. Attach to your IAM user

### 4. Configure AWS Profile

```bash
# Configure the 'aws' profile (or choose your own name)
aws configure --profile aws

# Enter when prompted:
AWS Access Key ID: [Your access key ID]
AWS Secret Access Key: [Your secret access key]
Default region name: us-east-1
Default output format: json
```

### 5. Test Configuration

```bash
# Test basic access
aws sts get-caller-identity --profile aws

# Test EC2 access (should list Rocky Linux AMIs)
aws ec2 describe-images \
    --owners=679593333241 \
    --filters "Name=name,Values=Rocky-9-EC2-Base-9.*x86_64*" \
    --max-items 1 \
    --profile aws

# Test ECR access
aws ecr describe-repositories --profile aws
```

### 6. Create EC2 Key Pair

```bash
# Create key pair for build instances
aws ec2 create-key-pair \
    --key-name geoschem-builder-key \
    --profile aws \
    --query 'KeyMaterial' \
    --output text > ~/.ssh/geoschem-builder-key.pem

# Secure the key file
chmod 400 ~/.ssh/geoschem-builder-key.pem

# Verify key was created
aws ec2 describe-key-pairs \
    --key-names geoschem-builder-key \
    --profile aws
```

### 7. Create ECR Repository

```bash
# Create repository for container images
aws ecr create-repository \
    --repository-name geoschem \
    --profile aws

# Get the repository URI for configuration
aws ecr describe-repositories \
    --repository-names geoschem \
    --query 'repositories[0].repositoryUri' \
    --output text \
    --profile aws
```

## Configuration Files

Update the platform configuration with your settings:

### `config/build-matrix.yaml`
```yaml
aws:
  profile: "aws"  # Your AWS profile name
  region: "us-east-1"  # Your preferred region
  key_pair: "geoschem-builder-key"
  # These will be set by Terraform:
  security_group: "sg-xxxxxxxxx"  
  subnet_id: "subnet-xxxxxxxxx"

ecr_repository: "123456789012.dkr.ecr.us-east-1.amazonaws.com/geoschem"
```

Replace `123456789012` with your AWS account ID and `us-east-1` with your region.

## Troubleshooting

### Permission Denied Errors
```bash
# Check which user you're authenticated as
aws sts get-caller-identity --profile aws

# Verify IAM policies are attached
aws iam list-attached-user-policies --user-name your-username --profile aws
```

### Profile Issues
```bash
# List configured profiles
aws configure list-profiles

# Check profile configuration
aws configure list --profile aws

# Reconfigure if needed
aws configure --profile aws
```

### Region Availability
```bash
# Check if Rocky Linux 9 is available in your region
aws ec2 describe-images \
    --owners=679593333241 \
    --filters "Name=name,Values=Rocky-9-*" \
    --query 'Images[0:3].[Name,CreationDate]' \
    --profile aws

# If empty, try a different region like us-west-2
aws ec2 describe-images \
    --owners=679593333241 \
    --filters "Name=name,Values=Rocky-9-*" \
    --region us-west-2 \
    --query 'Images[0:3].[Name,CreationDate]' \
    --profile aws
```

## Security Best Practices

1. **Use minimal IAM permissions** (Option B above)
2. **Enable MFA** on your AWS account
3. **Rotate access keys** regularly
4. **Use separate AWS profiles** for different environments
5. **Never commit AWS credentials** to version control
6. **Use AWS IAM roles** when running on EC2 instances

## Cost Considerations

- **EC2 instances** for building containers (typically $0.10-0.40/hour)
- **ECR storage** for container images
- **S3 storage** for results and web hosting
- **Data transfer** costs for large datasets

Monitor usage through AWS Cost Explorer and set up billing alerts.

## Next Steps

After completing AWS setup:

1. Deploy infrastructure with Terraform
2. Build your first container
3. Run a test simulation

See the main [README.md](../README.md) for platform usage instructions.