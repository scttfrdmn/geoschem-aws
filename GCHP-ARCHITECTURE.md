# GeosChem High Performance (GCHP) Architecture Design

This document outlines the architecture for supporting both GeosChem Classic and GeosChem High Performance (GCHP) with MPI capabilities on AWS.

## GeosChem Build Targets

### 1. GeosChem Classic
- **Single-node execution**
- **OpenMP parallelization**
- **Traditional grid-based simulations**
- **Good for**: Development, small-scale runs, validation

### 2. GeosChem High Performance (GCHP)
- **Multi-node MPI execution**  
- **Cube-sphere grid on FV3 dynamical core**
- **Scalable to 1000+ cores**
- **Good for**: Production runs, high-resolution simulations, operational forecasting

## MPI Implementation Strategy

### Supported MPI Libraries
```yaml
mpi_implementations:
  openmpi:
    version: "5.0.1"
    architectures: [x86_64, arm64]
    aws_support: "Native HPC networking (EFA, SR-IOV)"
    
  intel_mpi:
    version: "2021.10" 
    architectures: [x86_64]
    aws_support: "Optimized for Intel instances"
    
  mpich:
    version: "4.1.2"
    architectures: [x86_64, arm64]  
    aws_support: "Lightweight, broad compatibility"
```

### Container Design Philosophy
- **Base Layer**: OS + compilers + MPI runtime
- **GeosChem Layer**: GEOS-Chem source + dependencies (GEOS-ESM, HEMCO, etc.)
- **GCHP Layer**: FV3 + MAPL + GCHP-specific components
- **Application Layer**: Run scripts + configuration management

## AWS Instance Strategy

### Development & Testing
```yaml
single_node:
  instance_types: [c5.2xlarge, c6g.2xlarge]
  cores: 8
  memory: 16GB
  use_cases: ["Classic GeosChem", "GCHP development"]
```

### Production GCHP
```yaml
multi_node_hpc:
  instance_types: [hpc7a.48xlarge, c6i.24xlarge, c6g.16xlarge]
  cores: 96-192 per node
  memory: 192-384GB per node
  networking: "EFA (Elastic Fabric Adapter) for low-latency MPI"
  placement_groups: "cluster" # for maximum inter-node bandwidth
```

### AWS Batch Integration
```yaml
batch_environments:
  classic:
    compute_type: "EC2"
    instance_types: [c5.large, c5.xlarge, c5.2xlarge]
    
  gchp_small:
    compute_type: "EC2" 
    instance_types: [hpc7a.12xlarge, c6i.8xlarge]
    placement_strategy: "spread"
    
  gchp_production:
    compute_type: "EC2"
    instance_types: [hpc7a.48xlarge]
    placement_strategy: "cluster"
    networking: "EFA"
```

## Container Architecture

### Multi-Stage Build Strategy

#### Stage 1: Base MPI Runtime
```dockerfile
FROM rockylinux:9 as mpi-base
# Install compiler toolchain
# Install MPI library (OpenMPI/Intel MPI/MPICH)  
# Configure MPI for containerized execution
# Install AWS networking optimizations
```

#### Stage 2: GeosChem Dependencies  
```dockerfile
FROM mpi-base as geoschem-deps
# Install GEOS-ESM libraries
# Install HEMCO
# Install NetCDF/HDF5 with parallel I/O
# Install scientific libraries (BLAS, LAPACK, etc.)
```

#### Stage 3: GeosChem Classic
```dockerfile
FROM geoschem-deps as geoschem-classic
# Build GeosChem classic configuration
# Install input data management tools
# Configure for single-node execution
```

#### Stage 4: GCHP High Performance
```dockerfile
FROM geoschem-deps as gchp-production
# Build FV3 dynamical core
# Build MAPL (Modeling Analysis and Prediction Layer)
# Build GCHP with MPI support
# Configure for multi-node execution
# Install job scheduling integration
```

### Final Container Tags
```yaml
container_matrix:
  classic:
    - "geoschem:classic-gcc13-x86_64"
    - "geoschem:classic-gcc13-arm64"
    
  gchp:
    - "geoschem:gchp-openmpi-gcc13-x86_64"
    - "geoschem:gchp-openmpi-gcc13-arm64" 
    - "geoschem:gchp-intelmpi-intel2024-x86_64"
    - "geoschem:gchp-mpich-gcc13-x86_64"
    - "geoschem:gchp-mpich-gcc13-arm64"
```

## Build Configuration Matrix

### Compiler + MPI Combinations
```yaml
build_matrix:
  x86_64:
    gcc13:
      mpi: [openmpi, mpich]
      targets: [classic, gchp]
      
    intel2024:
      mpi: [intelmpi, openmpi]  
      targets: [classic, gchp]
      optimizations: ["AVX-512", "Intel MKL"]
      
    aocc4:
      mpi: [openmpi]
      targets: [classic, gchp] 
      optimizations: ["AMD Zen optimizations"]
      
  arm64:
    gcc13:
      mpi: [openmpi, mpich]
      targets: [classic, gchp]
      optimizations: ["ARM Neon", "Graviton3 tuning"]
```

## Input Data Management

### GEOS-Chem Input Data
- **Size**: ~200GB-2TB depending on simulation
- **Storage**: AWS S3 with intelligent tiering
- **Access**: Mountable via s3fs or direct S3 API
- **Caching**: Instance-local NVMe for active datasets

### Met Fields & Emissions  
- **Real-time**: Integration with AWS weather data services
- **Historical**: S3 archival with lifecycle policies
- **Regional**: Subset data management for cost optimization

## Networking & Performance

### MPI Network Optimization
```yaml
network_stack:
  efa_support: true  # Elastic Fabric Adapter
  sr_iov: true       # Single Root I/O Virtualization
  placement_groups: "cluster"  # 10 Gbps within group
  
mpi_tuning:
  openmpi:
    btl: "^openib"  # Use AWS networking instead of InfiniBand emulation
    pml: "ucx"      # Use UCX for high-performance messaging
    
  intel_mpi:
    fabric: "shm:ofi" # Shared memory + libfabric
    provider: "efa"   # Use AWS EFA directly
```

### I/O Optimization  
- **Parallel NetCDF**: Enable collective I/O operations
- **Striping**: Distribute files across multiple EBS volumes
- **Caching**: Instance store NVMe for temporary files

## Deployment Patterns

### Development Workflow
```bash
# Single-node GCHP development
docker run --rm -v $(pwd):/workspace geoschem:gchp-openmpi-gcc13-x86_64 \
  gchp-dev-run --cores 8 --resolution C48

# Multi-node testing (via AWS Batch)
aws batch submit-job --job-definition gchp-test-job \
  --job-queue gchp-cluster-queue \
  --parameters resolution=C180,nodes=4,cores_per_node=48
```

### Production Deployment
```bash  
# Large-scale GCHP simulation
aws batch submit-job --job-definition gchp-production-job \
  --job-queue gchp-hpc-queue \
  --parameters resolution=C360,nodes=16,cores_per_node=96,runtime=48h
```

## Performance Targets

### GeosChem Classic
- **Single c5.2xlarge**: ~30-50 simulation days per wall-clock hour
- **Cost**: ~$0.34/hour for instance + storage

### GCHP Production  
- **16x hpc7a.48xlarge**: ~500-1000 simulation days per wall-clock hour
- **Cost**: ~$50/hour for instances, but 20-40x faster throughput
- **Cost per simulation day**: Potentially lower than Classic for large runs

## Next Steps

1. **Implement base MPI container** with OpenMPI 5.0.1
2. **Build GEOS-Chem + HEMCO dependencies** with parallel NetCDF
3. **Create GCHP build pipeline** with FV3 + MAPL
4. **Test multi-node MPI execution** on AWS Batch
5. **Optimize for EFA networking** and Graviton3 performance
6. **Implement input data management** with S3 integration

---

**Target Timeline**: 
- Phase 1 (MPI containers): 2-3 weeks  
- Phase 2 (GCHP integration): 4-6 weeks
- Phase 3 (Production optimization): 2-4 weeks

**Priority**: Focus on OpenMPI + GCC13 combination first (broadest compatibility), then Intel MPI for x86_64 performance optimization.