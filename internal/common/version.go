package common

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// Version information following Semantic Versioning 2.0.0
const (
    // Version is the current version of the application
    Version = "0.1.0"
    
    // Name is the application name
    Name = "GeosChem AWS Platform"
    
    // Copyright information
    Copyright = "Copyright (c) 2025 Scott Friedman"
    
    // License information
    License = "MIT License"
)

// GetVersion returns the current version
func GetVersion() string {
    return Version
}

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
    return fmt.Sprintf("%s v%s\n%s\nLicensed under %s",
        Name, Version, Copyright, License)
}

// ReadVersionFromFile attempts to read version from VERSION file
func ReadVersionFromFile() (string, error) {
    // Try to find VERSION file in project root
    wd, err := os.Getwd()
    if err != nil {
        return Version, nil // fallback to embedded version
    }
    
    // Look for VERSION file in current directory or parent directories
    for i := 0; i < 5; i++ { // max 5 levels up
        versionFile := filepath.Join(wd, "VERSION")
        if data, err := os.ReadFile(versionFile); err == nil {
            version := strings.TrimSpace(string(data))
            if version != "" {
                return version, nil
            }
        }
        // Move up one directory
        parent := filepath.Dir(wd)
        if parent == wd {
            break // reached root
        }
        wd = parent
    }
    
    return Version, nil // fallback to embedded version
}