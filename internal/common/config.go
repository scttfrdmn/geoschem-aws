package common

import (
    "fmt"
    "os"
    "gopkg.in/yaml.v3"
)

// AWSConfig holds AWS-specific configuration
type AWSConfig struct {
    Profile       string `yaml:"profile"`
    Region        string `yaml:"region"`
    KeyPair       string `yaml:"key_pair"`
    SecurityGroup string `yaml:"security_group"`
    SubnetID      string `yaml:"subnet_id"`
}

// BatchConfig holds AWS Batch configuration
type BatchConfig struct {
    ComputeEnvironment string `yaml:"compute_environment"`
    JobQueue          string `yaml:"job_queue"`
    JobDefinition     string `yaml:"job_definition"`
}

// CompilerConfig holds compiler-specific configuration
type CompilerConfig struct {
    Version    string   `yaml:"version"`
    MPIOptions []string `yaml:"mpi_options"`
}

// ArchConfig holds architecture-specific configuration
type ArchConfig struct {
    InstanceType string                    `yaml:"instance_type"`
    Compilers    map[string]CompilerConfig `yaml:"compilers"`
}

// BuildConfig holds the complete build matrix configuration
type BuildConfig struct {
    AWS           AWSConfig             `yaml:"aws"`
    Batch         BatchConfig           `yaml:"batch"`
    Architectures map[string]ArchConfig `yaml:"architectures"`
    MPIVersions   map[string]string     `yaml:"mpi_versions"`
    ECRRepository string                `yaml:"ecr_repository"`
}

// LoadBuildConfig loads configuration from YAML file
func LoadBuildConfig(configFile string) (*BuildConfig, error) {
    data, err := os.ReadFile(configFile)
    if err != nil {
        return nil, fmt.Errorf("reading config file: %w", err)
    }
    
    var config BuildConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("parsing config file: %w", err)
    }
    
    // Validate required fields
    if config.AWS.Profile == "" {
        return nil, fmt.Errorf("AWS profile is required")
    }
    if config.AWS.Region == "" {
        return nil, fmt.Errorf("AWS region is required")
    }
    
    return &config, nil
}

// LoadAWSConfig loads AWS-specific configuration from YAML file
func LoadAWSConfig(configFile string) (*AWSConfig, error) {
    data, err := os.ReadFile(configFile)
    if err != nil {
        return nil, fmt.Errorf("reading AWS config file: %w", err)
    }
    
    var configStruct struct {
        AWS   AWSConfig   `yaml:"aws"`
        Batch BatchConfig `yaml:"batch"`
    }
    
    if err := yaml.Unmarshal(data, &configStruct); err != nil {
        return nil, fmt.Errorf("parsing AWS config file: %w", err)
    }
    
    return &configStruct.AWS, nil
}