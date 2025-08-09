package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/scttfrdmn/geoschem-aws/internal/builder"
    "github.com/scttfrdmn/geoschem-aws/internal/common"
)

func main() {
    var (
        profile    = flag.String("profile", "aws", "AWS profile to use")
        region     = flag.String("region", "", "AWS region (overrides config file)")
        arch       = flag.String("arch", "", "Architecture: x86_64 or arm64")
        compiler   = flag.String("compiler", "", "Compiler: intel2024, gcc13, aocc4")
        mpi        = flag.String("mpi", "", "MPI: intelmpi, openmpi, mpich")
        buildAll   = flag.Bool("build-all", false, "Build all combinations for specified arch")
        buildMatrix = flag.Bool("build-matrix", false, "Build complete matrix")
        configFile = flag.String("config", "config/build-matrix.yaml", "Config file path")
        version    = flag.Bool("version", false, "Show version information")
        checkQuotas = flag.Bool("check-quotas", false, "Check AWS quotas before building")
        recommendInstance = flag.Bool("recommend-instance", false, "Get instance type recommendations")
        gridRes = flag.String("grid-resolution", "4x5", "Grid resolution (4x5, 2x2.5, 0.5x0.625)")
        speciesCount = flag.Int("species-count", 100, "Number of chemical species")
        budget = flag.Float64("budget-per-hour", 0, "Maximum cost per hour (0 = no limit)")
        priority = flag.String("priority", "balanced", "Optimization priority (cost, performance, balanced)")
    )
    flag.Parse()

    ctx := context.Background()

    // Handle version flag
    if *version {
        fmt.Println(common.GetVersionInfo())
        os.Exit(0)
    }

    // Handle instance recommendations
    if *recommendInstance {
        // Use default region if not specified
        recommendRegion := *region
        if recommendRegion == "" {
            recommendRegion = "us-west-2" // Default region
        }
        
        cfg, err := config.LoadDefaultConfig(ctx, 
            config.WithSharedConfigProfile(*profile),
            config.WithRegion(recommendRegion),
        )
        if err != nil {
            log.Fatalf("Failed to load AWS config: %v", err)
        }

        selector := common.NewInstanceSelector(cfg, recommendRegion)
        workload := common.WorkloadProfile{
            GridResolution: *gridRes,
            SpeciesCount:   *speciesCount,
            BudgetPerHour:  *budget,
            Priority:       *priority,
            Architecture:   "any", // Allow both x86_64 and ARM64
        }

        recommendations, err := selector.GetRecommendations(ctx, workload)
        if err != nil {
            log.Fatalf("Failed to get recommendations: %v", err)
        }

        fmt.Println(common.FormatRecommendations(recommendations, workload))
        os.Exit(0)
    }
    
    // Load configuration
    config, err := common.LoadBuildConfig(*configFile)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Override AWS profile and region if specified
    if *profile != "" {
        config.AWS.Profile = *profile
    }
    if *region != "" {
        config.AWS.Region = *region
    }

    fmt.Printf("%s v%s\n", common.Name, common.GetVersion())
    fmt.Printf("Using AWS Profile: %s, Region: %s\n", config.AWS.Profile, config.AWS.Region)

    // Initialize builder
    b, err := builder.New(ctx, config.AWS.Profile, config.AWS.Region)
    if err != nil {
        log.Fatalf("Failed to initialize builder: %v", err)
    }

    // Check quotas if requested or before major builds
    if *checkQuotas || *buildMatrix {
        fmt.Println("\n🔍 Checking AWS quotas...")
        if err := b.CheckQuotas(ctx); err != nil {
            log.Printf("Warning: Could not check quotas: %v", err)
            fmt.Println("Continuing with build (quota check failed)...")
        }
        fmt.Println()
    }

    switch {
    case *buildMatrix:
        fmt.Println("Building complete matrix...")
        err = b.BuildMatrix(ctx, config)
    case *buildAll:
        if *arch == "" {
            log.Fatal("--arch required with --build-all")
        }
        fmt.Printf("Building all combinations for %s...\n", *arch)
        err = b.BuildAllForArch(ctx, config, *arch)
    default:
        if *arch == "" || *compiler == "" || *mpi == "" {
            log.Fatal("--arch, --compiler, and --mpi required for single build")
        }
        fmt.Printf("Building single combination: %s-%s-%s\n", *arch, *compiler, *mpi)
        err = b.BuildSingle(ctx, config, *arch, *compiler, *mpi)
    }

    if err != nil {
        log.Fatalf("Build failed: %v", err)
    }

    fmt.Println("Build completed successfully!")
}