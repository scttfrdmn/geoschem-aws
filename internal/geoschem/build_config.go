package geoschem

import (
	"fmt"
	"strings"

	"github.com/scttfrdmn/geoschem-aws/internal/docker"
)

// GeosChem build configurations for different compiler and architecture combinations
type BuildMatrix struct {
	Configurations []BuildConfiguration `yaml:"configurations"`
}

type BuildConfiguration struct {
	Name         string            `yaml:"name"`
	Architecture string            `yaml:"architecture"`
	Compiler     string            `yaml:"compiler"`
	BaseImage    string            `yaml:"base_image"`
	BuildArgs    map[string]string `yaml:"build_args"`
	Description  string            `yaml:"description"`
}

// GetStandardBuildConfigs returns standard GeosChem build configurations
func GetStandardBuildConfigs() []BuildConfiguration {
	return []BuildConfiguration{
		{
			Name:         "geoschem-gcc-x86_64",
			Architecture: "x86_64",
			Compiler:     "gcc13",
			BaseImage:    "rockylinux:9",
			BuildArgs: map[string]string{
				"COMPILER":     "gcc",
				"COMPILER_VERSION": "13",
				"ARCHITECTURE": "x86_64",
				"SPACK_SPEC":   "geos-chem@14.4.3 %gcc@13.2.0",
			},
			Description: "GeosChem with GCC 13 on x86_64",
		},
		{
			Name:         "geoschem-intel-x86_64",
			Architecture: "x86_64", 
			Compiler:     "intel2024",
			BaseImage:    "rockylinux:9",
			BuildArgs: map[string]string{
				"COMPILER":     "intel",
				"COMPILER_VERSION": "2024.0",
				"ARCHITECTURE": "x86_64",
				"SPACK_SPEC":   "geos-chem@14.4.3 %intel@2024.0.0",
			},
			Description: "GeosChem with Intel Compiler 2024 on x86_64",
		},
		{
			Name:         "geoschem-gcc-arm64",
			Architecture: "arm64",
			Compiler:     "gcc13",
			BaseImage:    "rockylinux:9",
			BuildArgs: map[string]string{
				"COMPILER":     "gcc",
				"COMPILER_VERSION": "13",
				"ARCHITECTURE": "arm64",
				"SPACK_SPEC":   "geos-chem@14.4.3 %gcc@13.2.0",
			},
			Description: "GeosChem with GCC 13 on ARM64/Graviton",
		},
		{
			Name:         "geoschem-aocc-x86_64",
			Architecture: "x86_64",
			Compiler:     "aocc4",
			BaseImage:    "rockylinux:9",
			BuildArgs: map[string]string{
				"COMPILER":     "aocc",
				"COMPILER_VERSION": "4.0.0",
				"ARCHITECTURE": "x86_64", 
				"SPACK_SPEC":   "geos-chem@14.4.3 %aocc@4.0.0",
			},
			Description: "GeosChem with AMD AOCC 4 on x86_64",
		},
	}
}

// ToDockerBuildConfig converts a GeosChem build config to Docker build config
func (bc *BuildConfiguration) ToDockerBuildConfig(sourceRepo, sourceBranch, imageTag string) *docker.BuildConfig {
	return &docker.BuildConfig{
		SourceRepo:    sourceRepo,
		SourceBranch:  sourceBranch,
		DockerfileDir: "docker", // Assume Dockerfile is in docker/ subdirectory
		ImageName:     bc.Name,
		ImageTag:      imageTag,
		Architecture:  bc.Architecture,
		BuildArgs:     bc.BuildArgs,
	}
}

// GetBuildConfigByName returns a build configuration by name
func GetBuildConfigByName(name string) (*BuildConfiguration, error) {
	configs := GetStandardBuildConfigs()
	
	for _, config := range configs {
		if config.Name == name {
			return &config, nil
		}
	}
	
	return nil, fmt.Errorf("build configuration '%s' not found", name)
}

// ListAvailableConfigs returns a formatted list of available configurations
func ListAvailableConfigs() string {
	configs := GetStandardBuildConfigs()
	var result strings.Builder
	
	result.WriteString("Available GeosChem Build Configurations:\n\n")
	
	for _, config := range configs {
		result.WriteString(fmt.Sprintf("• %s\n", config.Name))
		result.WriteString(fmt.Sprintf("  Architecture: %s\n", config.Architecture))
		result.WriteString(fmt.Sprintf("  Compiler: %s\n", config.Compiler))
		result.WriteString(fmt.Sprintf("  Description: %s\n", config.Description))
		result.WriteString("\n")
	}
	
	return result.String()
}

// ValidateConfiguration validates a build configuration
func (bc *BuildConfiguration) Validate() error {
	if bc.Name == "" {
		return fmt.Errorf("configuration name is required")
	}
	
	if bc.Architecture != "x86_64" && bc.Architecture != "arm64" {
		return fmt.Errorf("architecture must be x86_64 or arm64, got: %s", bc.Architecture)
	}
	
	if bc.BaseImage == "" {
		return fmt.Errorf("base image is required")
	}
	
	if bc.BuildArgs == nil {
		bc.BuildArgs = make(map[string]string)
	}
	
	// Ensure required build args are present
	requiredArgs := []string{"COMPILER", "ARCHITECTURE"}
	for _, arg := range requiredArgs {
		if _, exists := bc.BuildArgs[arg]; !exists {
			return fmt.Errorf("required build arg '%s' is missing", arg)
		}
	}
	
	return nil
}

// GetDockerfilePath returns the expected Dockerfile path for this configuration
func (bc *BuildConfiguration) GetDockerfilePath() string {
	return fmt.Sprintf("docker/Dockerfile.%s", bc.Compiler)
}