# Instance Type Optimization for GeosChem Platform

This guide provides research-backed recommendations for AWS instance types optimized for GeosChem atmospheric chemistry simulations.

## GeosChem Computational Characteristics

### Workload Analysis
GeosChem is a **CPU-intensive, memory-moderate, I/O-light** atmospheric chemistry transport model with these characteristics:

- **CPU-bound**: Heavy floating-point operations, chemical kinetics calculations
- **Memory usage**: 2-16 GB typical, scales with grid resolution and species count
- **I/O patterns**: Sequential reads (meteorology), periodic writes (output)
- **Network**: Minimal (mainly for data access from S3)
- **Storage**: Temporary scratch space for input/output files

### Parallelization
- **Shared memory**: OpenMP threading within nodes
- **Distributed memory**: MPI across nodes for larger simulations
- **Optimal core counts**: 4-36 cores per simulation (diminishing returns beyond)
- **Memory per core**: 1-4 GB typical

## Container Building vs. Runtime Requirements

### Container Building Phase
**Characteristics:**
- CPU-intensive compilation (gcc, Intel, AMD compilers)
- Memory: 4-8 GB sufficient
- Duration: 30-90 minutes per container
- Single-threaded to moderate parallelism

**Recommended instances:**
- **c5.2xlarge** (8 vCPU, 16 GB) - $0.34/hour
- **c6g.2xlarge** (8 vCPU, 16 GB) - $0.27/hour (ARM64, 20% cost savings)

### Simulation Runtime Phase  
**Characteristics:**
- CPU-intensive atmospheric chemistry calculations
- Memory: Scales with grid resolution (2-32 GB typical)
- Duration: Hours to days
- Benefits from CPU optimization

## Instance Type Recommendations

### Tier 1: Recommended (Best Value)

#### For Most Users (4x5 degree, moderate species)
**c5.xlarge** - 4 vCPU, 8 GB RAM - $0.17/hour
- **Use case**: Standard global simulations
- **Memory**: 2 GB per core (optimal for most runs)
- **Cost**: Most economical for typical workloads
- **Performance**: Balanced CPU/memory ratio

**c6g.xlarge** - 4 vCPU, 8 GB RAM - $0.136/hour (20% savings)
- **Use case**: Same as c5.xlarge but ARM64
- **Requirement**: Ensure GeosChem ARM64 compatibility
- **Benefit**: 20% cost reduction

#### For High-Resolution Runs (2x2.5 degree, many species)
**c5.2xlarge** - 8 vCPU, 16 GB RAM - $0.34/hour
- **Use case**: High-resolution global or detailed regional
- **Memory**: Sufficient for most high-res scenarios
- **Performance**: Good parallel efficiency

**c6g.2xlarge** - 8 vCPU, 16 GB RAM - $0.27/hour
- **Use case**: Same as c5.2xlarge but ARM64
- **Benefit**: $0.07/hour savings

### Tier 2: High-Performance (Premium)

#### For Memory-Intensive Simulations
**r5.2xlarge** - 8 vCPU, 64 GB RAM - $0.50/hour
- **Use case**: Ultra-high resolution, extensive chemistry
- **Memory**: 8 GB per core for memory-bound workloads
- **Trade-off**: Higher cost for memory headroom

**r6g.2xlarge** - 8 vCPU, 64 GB RAM - $0.40/hour
- **Use case**: Same as r5.2xlarge but ARM64
- **Benefit**: 20% cost savings

#### For Compute-Intensive Chemistry
**c5.4xlarge** - 16 vCPU, 32 GB RAM - $0.68/hour
- **Use case**: Large-scale ensemble runs
- **Parallelism**: Good for highly parallel simulations
- **Efficiency**: May hit diminishing returns >12 cores

### Tier 3: Specialized Use Cases

#### For Development/Testing
**t3.medium** - 2 vCPU, 4 GB RAM - $0.042/hour
- **Use case**: Code testing, small test runs
- **Limitation**: Not suitable for production runs
- **Benefit**: Very low cost for development

#### For Extreme Scale
**c5.24xlarge** - 96 vCPU, 192 GB RAM - $4.08/hour
- **Use case**: Multiple concurrent simulations
- **Strategy**: Run multiple independent jobs
- **Warning**: High cost, ensure full utilization

## Performance vs. Cost Analysis

### Cost-Performance Ranking ($/core/hour)

| Instance Type | $/core/hour | Use Case | Efficiency |
|---------------|-------------|----------|------------|
| **t3.medium** | $0.021 | Development | Good for testing |
| **c6g.xlarge** | $0.034 | Standard ARM64 | ⭐ **Best value** |
| **c5.xlarge** | $0.043 | Standard x86_64 | ⭐ **Most compatible** |
| **c6g.2xlarge** | $0.034 | High-res ARM64 | ⭐ **Best value scaled** |
| **c5.2xlarge** | $0.043 | High-res x86_64 | Good balance |
| **r6g.2xlarge** | $0.050 | Memory-intensive | Memory optimization |
| **r5.2xlarge** | $0.063 | Memory-intensive | Traditional choice |

### Memory Cost Analysis ($/GB/hour)

| Instance Type | $/GB/hour | Memory/Core | Optimal For |
|---------------|-----------|-------------|-------------|
| **c5.xlarge** | $0.021 | 2 GB | Standard chemistry |
| **c6g.xlarge** | $0.017 | 2 GB | ⭐ **Best memory value** |
| **r6g.2xlarge** | $0.006 | 8 GB | Large datasets |
| **r5.2xlarge** | $0.008 | 8 GB | High-memory needs |

## Regional Instance Availability

### High Availability Regions
**us-east-1 (N. Virginia)**
- All instance types available
- Lowest prices typically
- High quota limits

**us-west-2 (Oregon)**
- All instance types available  
- Slight price premium (~5%)
- Good quota limits

**eu-west-1 (Ireland)**
- All instance types available
- EU data residency compliance
- Similar pricing to us-east-1

### Limited Availability
- **Graviton instances (c6g, r6g)**: Not available in all regions
- **Latest generations**: May not be available in newer regions

## Spot Instance Recommendations

### Spot Pricing (Typical 60-70% savings)
- **c5.xlarge**: ~$0.05/hour (vs $0.17 on-demand)
- **c5.2xlarge**: ~$0.10/hour (vs $0.34 on-demand)
- **Risk**: Interruption possible (2-minute warning)

### Spot Best Practices
1. **Use for non-critical workloads** 
2. **Implement checkpointing** in long simulations
3. **Mix instance types** to reduce interruption risk
4. **Monitor spot pricing trends** for optimal timing

## Instance Selection Guide

### Decision Matrix
```
Grid Resolution → Instance Size
Species Count ↓

         4x5°    2x2.5°   0.5x0.625°
<100     t3.med  c5.xl    c5.2xl
100-200  c5.xl   c5.2xl   r5.2xl  
200+     c5.2xl  r5.2xl   r5.4xl
```

### Quick Selection Rules
1. **Starting out**: c5.xlarge or c6g.xlarge
2. **High resolution**: c5.2xlarge or c6g.2xlarge  
3. **Memory issues**: r5.2xlarge or r6g.2xlarge
4. **Cost-sensitive**: Use Graviton (c6g, r6g) instances
5. **Development**: t3.medium for testing

## Performance Testing Recommendations

**To get definitive performance data, we recommend benchmarking with your specific GeosChem configuration:**

### Benchmark Test Plan
1. **Standard benchmark**: 4x5 degree, 1-month simulation
2. **Test matrix**: 3-4 instance types
3. **Metrics**: Runtime, cost, memory usage
4. **Configurations**: Different compiler/MPI combinations

### Suggested Benchmark Instances
- **c5.xlarge** (baseline)
- **c6g.xlarge** (ARM64 comparison)
- **c5.2xlarge** (scaling test)
- **r5.2xlarge** (memory comparison)

### Benchmark Script
```bash
#!/bin/bash
# Benchmark different instance types
for instance in c5.xlarge c6g.xlarge c5.2xlarge r5.2xlarge; do
    echo "Testing $instance..."
    # Launch instance, run standard GeosChem benchmark
    # Measure: execution time, peak memory, cost
done
```

## Cost Optimization Strategies

### 1. Right-Sizing
- **Start small** (c5.xlarge) and scale up only if needed
- **Monitor CPU utilization** - aim for >80%
- **Check memory usage** - ensure not hitting swap

### 2. Scheduling
- **Use Spot instances** for fault-tolerant workloads
- **Schedule during low-demand periods** for better spot pricing
- **Use Reserved instances** for predictable workloads (1-3 year terms)

### 3. Architecture Choice
- **Prefer Graviton** (ARM64) when compatible - 20% savings
- **Test compatibility** with your specific GeosChem build
- **Monitor performance** - ensure no degradation

## Implementation in Platform

### Automatic Instance Selection
The platform should include intelligent instance selection:

```yaml
# config/instance-recommendations.yaml
instance_profiles:
  development:
    instance_type: t3.medium
    max_cost_per_hour: 0.10
  
  standard:
    instance_type: c5.xlarge
    fallback: c6g.xlarge
    max_cost_per_hour: 0.20
    
  high_performance:
    instance_type: c5.2xlarge  
    fallback: c6g.2xlarge
    max_cost_per_hour: 0.40
```

### Smart Recommendations
```bash
# Platform should analyze and recommend
go run cmd/builder/main.go --recommend-instance --grid-resolution 4x5 --species-count 150
# Output: "Recommended: c5.xlarge (cost: $0.17/hr, runtime: ~2 hours)"
```

## Next Steps

1. **Implement benchmark suite** to validate these recommendations
2. **Add instance type selection** to platform configuration
3. **Create cost calculator** for different scenarios  
4. **Monitor performance** and update recommendations
5. **Test ARM64 compatibility** with GeosChem builds

## Questions for Validation

**We should validate these recommendations through actual testing:**

1. **ARM64 Performance**: How does GeosChem perform on Graviton vs Intel?
2. **Memory Scaling**: At what point do we need r5/r6 instances?
3. **Compiler Differences**: Do Intel/GCC/AMD compilers have different optimal instances?
4. **I/O Patterns**: Do we need enhanced networking for large datasets?

**Recommendation**: Run a systematic benchmark across 4-5 instance types with standard GeosChem configurations to validate these theoretical recommendations.