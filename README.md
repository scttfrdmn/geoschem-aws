# GeosChem AWS Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://semver.org)

A comprehensive platform for running GeosChem atmospheric chemistry simulations on AWS using containerized builds and AWS Batch, now powered by **Rocky Linux 9**.

**Version**: 0.1.0 (following [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html))

## Features
- Pre-built GeosChem containers with multiple compiler/MPI combinations on Rocky Linux 9
- Simple web interface for job submission and monitoring
- Automated container building on native AWS instances using CIQ's official Rocky Linux 9 AMIs
- Integration with AWS Open Data for input datasets
- Configurable AWS profiles and regions

## Quick Start

1. **Setup AWS Profile**
   ```bash
   aws configure --profile aws
   ```

2. **Configure the Platform**
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

### Common Issues
- **AMI not found**: Ensure CIQ Rocky Linux 9 is available in your target region
- **Profile errors**: Verify AWS profile configuration with `aws sts get-caller-identity --profile aws`
- **Region mismatch**: Ensure ECR repository matches your build region

## Support

- Check CloudWatch logs for build details
- Monitor EC2 instances during container builds
- Verify IAM permissions for cross-service access

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Scott Friedman

## Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html). For the versions available, see the [CHANGELOG.md](CHANGELOG.md) file.