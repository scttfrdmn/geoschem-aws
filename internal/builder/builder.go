package builder

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/batch"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/ecr"
    
    "github.com/scttfrdmn/geoschem-aws/internal/common"
)

type Builder struct {
    ec2Client     *ec2.Client
    ecrClient     *ecr.Client
    quotaChecker  *common.QuotaChecker
    profile       string
    region        string
}

type BuildRequest struct {
    Architecture string
    Compiler     string
    MPI          string
    Tag          string
}

func New(ctx context.Context, profile string, region string) (*Builder, error) {
    cfg, err := config.LoadDefaultConfig(ctx, 
        config.WithSharedConfigProfile(profile),
        config.WithRegion(region),
    )
    if err != nil {
        return nil, fmt.Errorf("loading AWS config with profile %s and region %s: %w", profile, region, err)
    }

    return &Builder{
        ec2Client:    ec2.NewFromConfig(cfg),
        ecrClient:    ecr.NewFromConfig(cfg),
        quotaChecker: common.NewQuotaChecker(cfg, region),
        profile:      profile,
        region:       region,
    }, nil
}

func (b *Builder) BuildMatrix(ctx context.Context, config *common.BuildConfig) error {
    fmt.Printf("Building complete matrix in region %s...\n", b.region)
    
    for arch, archConfig := range config.Architectures {
        fmt.Printf("Processing architecture: %s\n", arch)
        if err := b.BuildAllForArch(ctx, config, arch); err != nil {
            return fmt.Errorf("building arch %s: %w", arch, err)
        }
    }
    
    return nil
}

func (b *Builder) BuildAllForArch(ctx context.Context, config *common.BuildConfig, arch string) error {
    archConfig, exists := config.Architectures[arch]
    if !exists {
        return fmt.Errorf("unknown architecture: %s", arch)
    }

    fmt.Printf("Building all combinations for %s in region %s...\n", arch, b.region)
    
    for compiler, compilerConfig := range archConfig.Compilers {
        for _, mpi := range compilerConfig.MPIOptions {
            fmt.Printf("Building: %s-%s-%s\n", arch, compiler, mpi)
            if err := b.BuildSingle(ctx, config, arch, compiler, mpi); err != nil {
                return fmt.Errorf("building %s-%s-%s: %w", arch, compiler, mpi, err)
            }
        }
    }
    
    return nil
}

func (b *Builder) BuildSingle(ctx context.Context, config *common.BuildConfig, arch, compiler, mpi string) error {
    tag := fmt.Sprintf("%s-%s", compiler, mpi)
    if arch == "arm64" {
        tag += "-arm64"
    }
    
    fmt.Printf("Building: %s (using Rocky Linux 9 in %s)\n", tag, b.region)
    
    buildReq := BuildRequest{
        Architecture: arch,
        Compiler:     compiler,
        MPI:          mpi,
        Tag:          tag,
    }
    
    // Launch EC2 instance
    instanceID, err := b.launchBuildInstance(ctx, config, arch)
    if err != nil {
        return fmt.Errorf("launching instance: %w", err)
    }
    
    defer func() {
        if err := b.terminateInstance(ctx, instanceID); err != nil {
            fmt.Printf("Warning: failed to terminate instance %s: %v\n", instanceID, err)
        }
    }()
    
    // Wait for instance to be ready
    if err := b.waitForInstance(ctx, instanceID); err != nil {
        return fmt.Errorf("waiting for instance: %w", err)
    }
    
    // Execute build
    if err := b.executeBuild(ctx, instanceID, buildReq, config); err != nil {
        return fmt.Errorf("executing build: %w", err)
    }
    
    fmt.Printf("Successfully built: %s\n", tag)
    return nil
}

func (b *Builder) executeBuild(ctx context.Context, instanceID string, buildReq BuildRequest, config *common.BuildConfig) error {
    // This is a placeholder for the actual build execution
    // In a real implementation, you would:
    // 1. Wait for the user data script to complete
    // 2. SSH into the instance or use SSM
    // 3. Build the Docker container
    // 4. Push to ECR
    
    fmt.Printf("Executing build on instance %s for %s...\n", instanceID, buildReq.Tag)
    
    // Simulate build time
    time.Sleep(30 * time.Second)
    
    fmt.Printf("Build execution completed for %s\n", buildReq.Tag)
    return nil
}

// CheckQuotas checks AWS service quotas relevant to the platform
func (b *Builder) CheckQuotas(ctx context.Context) error {
    report, err := b.quotaChecker.CheckGeoChemQuotas(ctx)
    if err != nil {
        return fmt.Errorf("checking quotas: %w", err)
    }

    report.PrintReport()

    // Check if any critical quotas need attention
    for _, quota := range report.Quotas {
        if quota.Status == "CRITICAL" {
            fmt.Printf("\n🚨 CRITICAL: %s quota is at %.1f%% usage\n", quota.QuotaName, quota.Usage)
            if quota.CanIncrease {
                fmt.Printf("💡 Consider requesting a quota increase for %s\n", quota.ServiceName)
                fmt.Printf("   Use: aws support create-case --service-code=\"%s\" ...\n", strings.ToLower(quota.ServiceName))
            }
        }
    }

    return nil
}