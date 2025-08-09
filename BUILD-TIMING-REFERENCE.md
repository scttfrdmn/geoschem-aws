# GeosChem AWS Platform - Build Timing Reference

This document provides expected build times for container builds across different architectures and configurations.

## Container Build Times

### x86_64 Architecture (Intel/AMD)
- **Total Build Time**: ~5 minutes
- **Instance Setup**: ~2-3 minutes
- **Container Build**: ~2 minutes
- **ECR Push**: ~1 minute
- **Instance Types**: c5.2xlarge, c5.4xlarge recommended

### ARM64 Architecture (Graviton)
- **Total Build Time**: ~8-10 minutes (basic), ~15-20 minutes (full production with MPI)
- **Instance Setup**: ~4-5 minutes (more packages required)
- **Container Build**: ~3-4 minutes (basic), ~8-12 minutes (with Spack dependencies)
- **ECR Push**: ~1 minute  
- **Instance Types**: c6g.xlarge, c6g.2xlarge recommended

### Production Containers with AWS Optimizations
- **x86_64 with AWS Binary Cache**: ~10-15 minutes (vs 30-60 minutes without)
- **ARM64 with AWS Binary Cache**: ~15-20 minutes (vs 45-90 minutes without)
- **AWS Binary Cache Speedup**: Up to 20x faster for common packages
- **Spack Dependencies**: NetCDF, HDF5, ESMF, MPI libraries from pre-built binaries

## Why ARM64 Takes Longer

### Package Dependencies
- **x86_64**: ~60 packages to install
- **ARM64**: ~92 packages to install
- ARM64 Rocky Linux has additional container runtime dependencies

### Network Factors
- Instance location relative to package repositories
- Graviton instances may use different network paths
- Regional variations in mirror availability

### Build Process Differences
- **Dependency Resolution**: ARM64 requires more complex dependency chains
- **Package Compilation**: Some packages compile during installation on ARM64
- **Container Registry**: ARM64 base images are typically larger

## Performance Recommendations

### For Development (Speed Priority)
```bash
# x86_64 - Fastest builds
--arch x86_64 --instance-type c5.xlarge

# ARM64 - Skip system updates for faster iteration
--arch arm64 --instance-type c6g.xlarge --skip-update
```

### For Production (Reliability Priority)
```bash
# x86_64 - Full build with updates
--arch x86_64 --instance-type c5.2xlarge

# ARM64 - Full build with updates
--arch arm64 --instance-type c6g.2xlarge
```

## Cost Considerations

### x86_64 Build Cost (us-west-2)
- **c5.2xlarge**: ~$0.34/hour
- **5-minute build**: ~$0.03 per build

### ARM64 Build Cost (us-west-2) 
- **c6g.2xlarge**: ~$0.27/hour (20% less than x86_64)
- **8-minute build**: ~$0.04 per build

**Note**: ARM64 takes longer but uses cheaper Graviton instances, resulting in similar cost per build.

## Build Time Optimization Tips

### 1. Skip System Updates for Development
```bash
# Saves ~1-2 minutes per build
PrepareInstance(ctx, true) // skipUpdate = true
```

### 2. Use Larger Instance Types for Batch Builds
```bash
# For multiple builds, use c5.4xlarge or c6g.4xlarge
# Parallel builds can share setup overhead
```

### 3. Regional Instance Placement
- **us-west-2**: Generally fastest for West Coast users
- **us-east-1**: Generally fastest for East Coast users
- Avoid cross-region builds unless necessary

### 4. Build Caching (Future Enhancement)
- Container layer caching can reduce build times by 50-70%
- Base image pre-pulling reduces network transfer time
- Spack build cache for compiled packages

## Troubleshooting Slow Builds

### If ARM64 builds take >15 minutes:
1. Check instance network connectivity
2. Verify Rocky Linux mirror availability
3. Consider different AWS region
4. Check for DNS resolution issues

### If x86_64 builds take >10 minutes:
1. Instance may be under-provisioned
2. Network throttling possible
3. Check ECR region settings
4. Verify IAM permissions for faster ECR access

---

**Last Updated**: August 9, 2025  
**Tested Configurations**: Rocky Linux 9, Podman 5.4.0, AWS ECR  
**Benchmark Instance Types**: c5.2xlarge (x86_64), c6g.2xlarge (ARM64)