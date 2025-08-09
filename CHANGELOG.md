# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial GeosChem AWS Platform implementation
- Rocky Linux 9 support with CIQ official AMIs
- Configurable AWS profile and region support
- Dynamic AMI lookup for latest Rocky Linux 9 images
- Multi-architecture support (x86_64 and ARM64/Graviton)
- Multiple compiler and MPI combinations support
- Docker containerization with Spack package management
- Automated container building on native AWS instances
- AWS Batch integration for job execution
- Web interface for job submission and monitoring
- Comprehensive configuration system
- Cost optimization features (Spot instances, ARM64 support)

### Changed
- Migrated from Amazon Linux 2 to Rocky Linux 9
- Updated package management from `yum` to `dnf`
- Changed default user from `ec2-user` to `rocky`
- Replaced Ubuntu base images with Rocky Linux 9 in containers

### Security
- Non-root container execution with dedicated `geoschem` user
- Enterprise-grade security updates from Rocky Linux 9
- Secure AWS credential handling via profiles

## [0.1.0] - 2025-01-09

### Added
- Initial alpha release of GeosChem AWS Platform
- Basic platform for running atmospheric chemistry simulations on AWS
- Rocky Linux 9 enterprise foundation with CIQ official AMIs
- Configurable AWS profile and region support
- Multi-architecture support (x86_64 and ARM64)
- Docker containerization with Spack
- Core build system and configuration
- MIT license and semantic versioning
- Full documentation and getting started guide

### Note
- This is an early alpha release for development and testing