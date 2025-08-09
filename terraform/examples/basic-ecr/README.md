# Basic ECR Module Test

## Prerequisites

### Official Terraform Installation
```bash
# Install official Terraform (recommended)
brew tap hashicorp/tap
brew install hashicorp/tap/terraform

# Or use older version that's already installed
# terraform --version should show 1.5+
```

### AWS Profile
Ensure you have the 'aws' profile configured:
```bash
aws configure --profile aws
aws sts get-caller-identity --profile aws
```

## Test the ECR Module

```bash
# Initialize and validate
terraform init
terraform validate

# Plan deployment
terraform plan

# Deploy (creates ECR repository)
terraform apply -auto-approve

# Test the repository
aws ecr describe-repositories --repository-names geoschem-test --profile aws --region us-west-2

# Clean up
terraform destroy -auto-approve
```

## What This Tests

✅ **ECR Repository Creation**: Creates a repository named `geoschem-test`
✅ **Lifecycle Policy**: Automatically expires old images (keeps last 10)
✅ **Repository Policy**: Allows your IAM user to push/pull images
✅ **Image Scanning**: Enables vulnerability scanning on push
✅ **Tagging**: Applies consistent tags for resource management

## Success Criteria

- [ ] `terraform init` completes without errors
- [ ] `terraform validate` shows "Success! The configuration is valid."
- [ ] `terraform plan` shows 3 resources to be created
- [ ] `terraform apply` creates ECR repository successfully
- [ ] AWS CLI can describe the repository
- [ ] `terraform destroy` cleans up all resources

## Next Steps

This basic ECR module is now ready for:
- Integration into the main platform infrastructure
- Use by container build processes
- Storage of GeosChem container images

The module supports:
- Custom repository names
- Multiple allowed principals
- Flexible tagging
- Production-ready policies