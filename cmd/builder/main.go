package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/your-org/geoschem-aws-platform/internal/builder"
    "github.com/your-org/geoschem-aws-platform/internal/common"
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
    )
    flag.Parse()

    // Handle version flag
    if *version {
        fmt.Println(common.GetVersionInfo())
        os.Exit(0)
    }

    ctx := context.Background()
    
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