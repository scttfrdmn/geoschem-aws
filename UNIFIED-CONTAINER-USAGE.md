# Unified GeosChem Container - Supporting both Classic and GCHP

This document describes how to use the unified GeosChem container that supports both Classic and High Performance (GCHP) execution modes in a single container image.

## Container Architecture

The unified container includes:
- **GeosChem Classic**: Traditional single-node execution with OpenMP
- **GCHP**: Multi-node MPI execution with FV3 cube-sphere grid
- **Multiple MPI implementations**: OpenMPI, Intel MPI, MPICH
- **Multiple compilers**: GCC, Intel, AOCC
- **Multi-architecture**: x86_64 and ARM64

## Basic Usage

### GeosChem Classic Mode

```bash
# Basic classic simulation
docker run --rm -v $(pwd):/workspace \
  geoschem:production-gcc13-openmpi-x86_64 \
  classic --simulation fullchem --resolution 4x5

# Classic with custom date range
docker run --rm -v $(pwd):/workspace \
  geoschem:production-gcc13-openmpi-x86_64 \
  classic --simulation fullchem --resolution 2x2.5 \
  --start-date 2019-01-01 --end-date 2019-01-31
```

### GCHP High Performance Mode

```bash
# Basic GCHP simulation  
docker run --rm -v $(pwd):/workspace \
  geoschem:production-gcc13-openmpi-x86_64 \
  gchp --simulation fullchem --resolution C48 --cores 24

# High-resolution GCHP simulation
docker run --rm -v $(pwd):/workspace \
  geoschem:production-gcc13-openmpi-x86_64 \
  gchp --simulation fullchem --resolution C180 --cores 96 \
  --start-date 2019-07-01 --end-date 2019-07-31
```

## AWS Deployment Examples

### Single-Node Development (Classic + Small GCHP)

```bash
# Launch c5.2xlarge for development
aws batch submit-job \
  --job-name geoschem-dev-classic \
  --job-queue geoschem-dev-queue \
  --job-definition geoschem-unified-dev \
  --parameters simulation=fullchem,resolution=4x5,mode=classic

# Small GCHP test on same instance
aws batch submit-job \
  --job-name geoschem-dev-gchp \
  --job-queue geoschem-dev-queue \
  --job-definition geoschem-unified-dev \
  --parameters simulation=fullchem,resolution=C48,mode=gchp,cores=8
```

### Production GCHP Multi-Node

```bash
# Large-scale GCHP on HPC instances
aws batch submit-job \
  --job-name geoschem-production-gchp \
  --job-queue geoschem-hpc-queue \
  --job-definition geoschem-unified-hpc \
  --parameters simulation=fullchem,resolution=C360,mode=gchp,cores=192,nodes=4
```

## Performance Recommendations

### Instance Type Selection

| **Use Case** | **Mode** | **Resolution** | **Instance Type** | **Cores** | **Cost/Hour** |
|--------------|----------|----------------|-------------------|-----------|---------------|
| Development | Classic | 4x5, 2x2.5 | c5.xlarge | 4 | $0.17 |
| Development | GCHP | C48, C90 | c5.2xlarge | 8 | $0.34 |
| Production | Classic | 0.5x0.625 | c5.4xlarge | 16 | $0.68 |
| Production | GCHP | C180 | c6i.8xlarge | 32 | $1.34 |
| HPC | GCHP | C360+ | hpc7a.48xlarge | 192 | $13.23 |

### ARM64 Graviton Optimization

```bash
# ARM64 offers ~20% cost savings
docker run --rm -v $(pwd):/workspace \
  geoschem:production-gcc13-openmpi-arm64 \
  gchp --simulation fullchem --resolution C180 --cores 64

# Graviton3 instances
# c6g.16xlarge: 64 cores, $2.18/hour (vs c6i.16xlarge $2.68/hour)
# c7g.16xlarge: 64 cores, $2.31/hour (latest generation)
```

## Volume Mounting & Data Management

### Input Data Structure
```bash
workspace/
├── config/          # Custom configuration files
├── data/            # Input meteorological and emission data
│   ├── GEOS_4x5/   # Classic resolution data
│   ├── GEOS_0.25x0.3125/  # High-resolution data
│   └── ExtData/     # GCHP external data
├── output/          # Simulation output
│   ├── classic_*/   # Classic run directories
│   └── gchp_*/      # GCHP run directories
└── benchmarks/      # Benchmark configurations
```

### GeosChem AWS Open Data Integration
The container automatically configures access to GeosChem's official AWS Open Data Archive (~20TB), eliminating the need to download/store data locally.

```bash
# Automatic AWS Open Data access (recommended)
docker run --rm -v $(pwd):/workspace \
  geoschem:production-gcc13-openmpi-x86_64 \
  classic --simulation fullchem --resolution 4x5
  # Data accessed directly from s3://gcgrid

# Configure data sources explicitly  
docker run --rm -v $(pwd):/workspace \
  geoschem:production-gcc13-openmpi-x86_64 \
  configure-data-sources --resolution 4x5 --mode classic

# List available datasets
docker run --rm geoschem:production-gcc13-openmpi-x86_64 \
  configure-data-sources --list
```

### Available Data on AWS Open Data Archive
- **Meteorological Data**: GEOS 4x5, 2x2.5, 0.5x0.625, 0.25x0.3125
- **Emissions (HEMCO)**: All standard emission inventories  
- **Chemistry**: CHEM_INPUTS, photolysis data, chemical mechanisms
- **GCHP ExtData**: High-resolution cube-sphere datasets
- **Benchmarks**: Reference datasets for validation

**Benefits of AWS Open Data Access:**
- No local storage requirements (saves ~50GB-8TB per simulation)
- Faster access than downloading (especially in AWS regions)
- Always up-to-date datasets  
- Zero data transfer costs when running in AWS

## Container Variants

### Available Tags
```bash
# x86_64 variants
geoschem:production-gcc13-openmpi-x86_64     # Most common
geoschem:production-intel2024-intelmpi-x86_64  # Intel optimized
geoschem:production-aocc4-openmpi-x86_64     # AMD optimized

# ARM64 variants  
geoschem:production-gcc13-openmpi-arm64      # Graviton optimized
geoschem:production-gcc13-mpich-arm64        # Alternative MPI

# Development variants
geoschem:dev-gcc13-openmpi-x86_64           # With debugging tools
```

### Choosing the Right Variant

**For Development:**
- `geoschem:production-gcc13-openmpi-x86_64` - Broadest compatibility
- `geoschem:dev-*` - Include gdb, valgrind, extra debugging

**For Production:**
- x86_64: `geoschem:production-intel2024-intelmpi-x86_64` - Maximum performance
- ARM64: `geoschem:production-gcc13-openmpi-arm64` - Cost optimization
- AMD: `geoschem:production-aocc4-openmpi-x86_64` - AMD-optimized

## Advanced Configuration

### Custom MPI Settings
```bash
# Customize MPI behavior for specific networking
docker run --rm -v $(pwd):/workspace \
  -e OMPI_MCA_btl=^openib \
  -e OMPI_MCA_pml=ucx \
  geoschem:production-gcc13-openmpi-x86_64 \
  gchp --simulation fullchem --resolution C180 --cores 48
```

### Multi-Container Scaling
```bash
# Run multiple containers for parameter studies
for resolution in C48 C90 C180; do
  docker run --rm -d \
    --name geoschem-${resolution} \
    -v $(pwd)/${resolution}:/workspace \
    geoschem:production-gcc13-openmpi-x86_64 \
    gchp --simulation fullchem --resolution ${resolution} --cores 24
done
```

### Debugging and Development
```bash
# Run with debugging enabled
docker run --rm -it \
  -v $(pwd):/workspace \
  geoschem:dev-gcc13-openmpi-x86_64 \
  classic --simulation fullchem --resolution 4x5 --debug

# Interactive debugging session
docker run --rm -it \
  -v $(pwd):/workspace \
  geoschem:dev-gcc13-openmpi-x86_64 \
  /bin/bash
```

## Performance Benchmarking

### Benchmark Suite Integration
```bash
# Run standard benchmarks
docker run --rm -v $(pwd):/workspace \
  geoschem:production-gcc13-openmpi-x86_64 \
  benchmark --suite gchp-standard --resolution C180 --cores 48

# Performance comparison across architectures
./scripts/benchmark-comparison.sh \
  --resolutions "C48,C90,C180" \
  --architectures "x86_64,arm64" \
  --instance-types "c5.2xlarge,c6g.2xlarge"
```

### Expected Performance

| **Mode** | **Resolution** | **Cores** | **Instance** | **Sim Days/Hour** | **Cost/Sim Day** |
|----------|----------------|-----------|--------------|-------------------|-------------------|
| Classic | 4x5 | 4 | c5.xlarge | 30-50 | $0.003-0.006 |
| Classic | 2x2.5 | 8 | c5.2xlarge | 15-25 | $0.014-0.023 |
| GCHP | C48 | 24 | c5.6xlarge | 100-200 | $0.005-0.010 |
| GCHP | C180 | 96 | c6i.24xlarge | 200-400 | $0.013-0.027 |
| GCHP | C360 | 192 | hpc7a.48xlarge | 400-800 | $0.017-0.033 |

---

**Next Steps:**
1. **Container Testing**: Validate both modes work correctly
2. **Benchmark Suite**: Implement comprehensive performance testing
3. **Multi-Architecture**: Test ARM64 vs x86_64 performance  
4. **AWS Integration**: Optimize for EFA networking and Batch scheduling
5. **Data Pipeline**: Integrate S3 input data management