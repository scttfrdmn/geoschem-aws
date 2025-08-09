# Development Guide

## Quick Start for v0.1.0 Development

### Immediate Priority: Make It Work 🛠️

The platform currently has excellent foundations but **doesn't actually work yet**. Here are the critical gaps to address first:

## Phase 1: Critical Infrastructure (Start Here)

### 1. Create Working Terraform Infrastructure

**Current State**: Configuration files exist, no actual Terraform
**Priority**: CRITICAL
**Effort**: 3-5 days

```bash
# Create these files:
mkdir terraform/modules
terraform/
├── main.tf                 # Entry point
├── modules/
│   ├── vpc/               # VPC with security groups  
│   ├── ecr/               # Container registry
│   ├── batch/             # Batch compute environment
│   └── iam/               # Roles and policies
└── examples/
    └── basic/             # Basic deployment example
```

**Test Criteria**: `terraform apply` successfully creates AWS infrastructure

### 2. Fix Docker Build Execution 

**Current State**: Stubbed with `time.Sleep()` 
**Priority**: CRITICAL  
**Effort**: 4-6 days

**Replace this:**
```go
// Current stub in internal/builder/builder.go:133
time.Sleep(30 * time.Second)
```

**With actual implementation:**
```go
func (b *Builder) executeBuild(ctx context.Context, instanceID string, buildReq BuildRequest, config *common.BuildConfig) error {
    // 1. Wait for instance ready + user data completion
    if err := b.waitForInstanceReady(ctx, instanceID); err != nil {
        return err
    }
    
    // 2. Connect via SSH
    sshClient, err := b.connectSSH(instanceID, config.AWS.KeyPair)
    if err != nil {
        return err
    }
    defer sshClient.Close()
    
    // 3. Execute Docker build
    buildCmd := b.generateBuildCommand(buildReq)
    if err := b.runCommand(sshClient, buildCmd); err != nil {
        return err
    }
    
    // 4. Push to ECR
    pushCmd := b.generatePushCommand(buildReq, config)  
    return b.runCommand(sshClient, pushCmd)
}
```

**Test Criteria**: Successfully builds and pushes real GeosChem container to ECR

### 3. Implement AWS Batch Job Execution

**Current State**: Missing entirely
**Priority**: CRITICAL
**Effort**: 3-4 days

**Create new package:**
```go
// internal/batch/batch.go
type BatchClient struct {
    client *batch.Client
}

func (bc *BatchClient) SubmitGeoChemJob(ctx context.Context, jobSpec JobSpec) (string, error) {
    // Create job definition if not exists
    // Submit job with parameters
    // Return job ID for monitoring
}
```

**Test Criteria**: Can submit and monitor actual GeosChem simulation jobs

## Phase 2: User Experience (Week 2)

### 4. Add Configuration Validation

**Current Issue**: Users can misconfigure and waste time/money
**Priority**: IMPORTANT
**Effort**: 1-2 days

```go
// internal/common/validator.go
func ValidateConfig(config *BuildConfig) error {
    // Check AWS credentials
    // Validate region availability  
    // Check ECR repository exists
    // Verify key pair exists
    // Test instance quotas
}
```

### 5. Create Setup Wizard

**Current Issue**: Configuration is complex and error-prone
**Priority**: IMPORTANT
**Effort**: 2-3 days

```bash
# cmd/setup/main.go
go run cmd/setup/main.go
# Interactive prompts for:
# - AWS profile selection
# - Region with availability checking
# - Instance type recommendations  
# - Budget preferences
# - Infrastructure deployment
```

## Phase 3: Testing & Validation (Week 3)

### 6. Integration Tests

**Current State**: No tests exist
**Priority**: IMPORTANT  
**Effort**: 2-3 days

```go
// tests/integration/
func TestFullWorkflow(t *testing.T) {
    // 1. Deploy infrastructure with Terraform
    // 2. Build container
    // 3. Submit job
    // 4. Monitor completion
    // 5. Clean up resources
}
```

### 7. Instance Type Benchmarking

**Current State**: Theoretical recommendations only
**Priority**: VALIDATION
**Effort**: 3-5 days

**Benchmark Plan**:
- Test c5.xlarge vs c6g.xlarge (x86_64 vs ARM64)
- Test memory scaling with different species counts
- Validate cost vs performance assumptions
- Update recommendations based on real data

## Quick Development Setup

> **Important**: All development and testing should use the 'aws' profile consistently. This ensures compatibility with the platform's default configuration and documentation examples.

### 1. Prerequisites Check
```bash
# Verify you have everything needed
go version          # Should be 1.21+
terraform --version # Should be 1.5+
aws --version      # Should be 2.x
docker --version   # Any recent version

# Test AWS access with 'aws' profile
aws sts get-caller-identity --profile aws
```

### 2. Development Environment
```bash
# Clone and setup
git clone <your-fork>
cd geoschem-aws

# Install dependencies  
go mod download
go mod tidy

# Validate existing code
go build ./...
go test ./...

# Test current functionality
go run cmd/builder/main.go --version
go run cmd/builder/main.go --recommend-instance --grid-resolution 4x5
```

### 3. Start with Terraform

**Why Start Here**: Everything else depends on AWS infrastructure

```bash
# Create basic Terraform structure
mkdir -p terraform/modules/{vpc,ecr,batch,iam}

# Start with ECR module (simplest)
cat > terraform/modules/ecr/main.tf << 'EOF'
resource "aws_ecr_repository" "geoschem" {
  name                 = var.repository_name
  image_tag_mutability = "MUTABLE"
  
  image_scanning_configuration {
    scan_on_push = true
  }
  
  lifecycle_policy {
    policy = jsonencode({
      rules = [{
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }]
    })
  }
}
EOF

# Test the module
cd terraform/modules/ecr
terraform init
terraform validate
```

## Development Workflow

### 1. Pick a Component
Start with the most critical gaps:
1. **Terraform infrastructure** (blocks everything else)
2. **Docker build execution** (core functionality)  
3. **Batch integration** (job execution)
4. **Configuration validation** (user experience)
5. **Integration tests** (quality assurance)

### 2. Test-Driven Development
```bash
# Write test first
func TestECRIntegration(t *testing.T) {
    // Test ECR repository creation and image push
}

# Implement functionality
# Run test until it passes
go test ./tests/integration -v
```

### 3. Validate with Real AWS

**Important**: Test with real but minimal AWS resources
- Use t3.nano instances for testing (cheapest)
- Clean up resources immediately after testing
- Set billing alerts to avoid surprises

### 4. Documentation as You Go
- Update README.md with any new requirements
- Add examples for new functionality  
- Document any AWS permissions needed
- Update troubleshooting guides

## Cost Management During Development

### 1. Use Minimal Resources
```bash
# For testing, use smallest instances
instance_type = "t3.nano"    # $0.0052/hour
volume_size   = 8            # Minimum EBS size
```

### 2. Immediate Cleanup
```bash
# Always clean up after testing
terraform destroy
aws batch cancel-job --job-id <job-id>
aws ec2 terminate-instances --instance-ids <instance-id>
```

### 3. Set Billing Alerts
- Create AWS Budget for $10/month development spend
- Set up CloudWatch billing alarms
- Monitor costs daily during active development

## Common Development Pitfalls

### 1. AWS Permissions
**Issue**: Subtle IAM permission errors
**Solution**: Start with AdministratorAccess, then narrow down

### 2. Regional Availability  
**Issue**: Services not available in all regions
**Solution**: Test in us-west-2, us-east-1, eu-west-1 initially

### 3. Quota Limits
**Issue**: Hitting quotas during testing
**Solution**: Use `--check-quotas` before any builds

### 4. Terraform State
**Issue**: Terraform state conflicts
**Solution**: Use local state for development, remote for shared work

## Ready to Start?

### Recommended First PR: Basic Terraform Infrastructure

```bash
# 1. Create branch
git checkout -b feature/basic-terraform

# 2. Create minimal ECR repository
# (See example above)

# 3. Test it works
terraform apply

# 4. Clean up
terraform destroy

# 5. Commit and push
git add terraform/
git commit -m "Add basic ECR Terraform module

- Create ECR repository for container storage
- Add lifecycle policy for cost management  
- Include proper tagging and scanning
- Test with terraform apply/destroy

Addresses critical gap: missing infrastructure deployment"

git push origin feature/basic-terraform
```

The platform has excellent foundations - now let's make it actually work! 🚀