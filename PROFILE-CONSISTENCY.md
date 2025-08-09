# AWS Profile Consistency Guide

## ⚠️ **Important: Always Use 'aws' Profile**

The GeosChem AWS Platform is designed around the **'aws' profile** for consistency. All documentation, examples, and default configurations use this profile name.

## Why 'aws' Profile?

1. **Consistency**: All documentation and examples work without modification
2. **Simplicity**: No profile name confusion across team members
3. **Default Behavior**: Platform defaults to 'aws' profile when not specified
4. **Integration Testing**: All tests assume 'aws' profile exists

## Setup Commands (Use These Exactly)

```bash
# Configure the required 'aws' profile
aws configure --profile aws

# Test the 'aws' profile  
aws sts get-caller-identity --profile aws

# All platform commands use 'aws' by default
go run cmd/builder/main.go --check-quotas
go run cmd/builder/main.go --recommend-instance --grid-resolution 4x5

# Explicitly specify 'aws' for different regions  
go run cmd/builder/main.go --profile aws --region us-east-1 --build-matrix
```

## Development & Testing

**✅ Always use 'aws' profile for:**
- Development and testing
- Documentation examples  
- Integration tests
- Benchmark validation
- Infrastructure deployment

**❌ Don't use these profile names:**
- `default` (may conflict with other tools)
- `production` (confusing in examples)
- `dev` (not standard across team)
- `test` (reserved for testing frameworks)

## Configuration Files

All config files should reference the 'aws' profile:

```yaml
# config/build-matrix.yaml
aws:
  profile: "aws"  # Required: Use 'aws' profile for consistency
  region: "us-west-2"
```

## Terraform Variables

When we add Terraform, use:

```hcl
# terraform/variables.tf
variable "aws_profile" {
  description = "AWS profile to use"
  type        = string
  default     = "aws"  # Default to 'aws' profile
}
```

## Testing Commands

All testing should use the 'aws' profile:

```bash
# Test current functionality
go run cmd/builder/main.go --version
go run cmd/builder/main.go --check-quotas --profile aws
go run cmd/builder/main.go --recommend-instance --profile aws

# Future infrastructure testing
terraform apply -var="aws_profile=aws"
terraform destroy -var="aws_profile=aws"

# Future integration tests
AWS_PROFILE=aws go test ./tests/integration/...
```

## Environment Variables

For automation and CI/CD:

```bash
# Set default profile for all AWS CLI commands
export AWS_PROFILE=aws

# Now all commands use 'aws' profile automatically
aws sts get-caller-identity
go run cmd/builder/main.go --check-quotas
```

## Multiple Environments

If you need different environments, use regions instead of profiles:

```bash
# Development in us-west-2 (default)
go run cmd/builder/main.go --profile aws --region us-west-2

# Testing in us-east-1
go run cmd/builder/main.go --profile aws --region us-east-1

# Different AWS accounts should use different named profiles
aws configure --profile aws-prod    # Production account
aws configure --profile aws-dev     # Development account
```

## Troubleshooting

**Problem**: "Profile 'aws' not found"
**Solution**: 
```bash
aws configure --profile aws
# Enter your credentials when prompted
```

**Problem**: "Permission denied" errors
**Solution**:
```bash
# Verify profile works
aws sts get-caller-identity --profile aws

# Check profile configuration  
aws configure list --profile aws
```

**Problem**: "Wrong region" errors
**Solution**:
```bash
# Override region while keeping 'aws' profile
go run cmd/builder/main.go --profile aws --region us-west-2
```

## Summary

- **Always use**: `aws configure --profile aws`
- **Always test with**: `aws sts get-caller-identity --profile aws`
- **Always run with**: `--profile aws` (or let it default)
- **Never use**: random profile names in documentation or examples

This ensures everyone can follow the same instructions and get consistent results! 🎯