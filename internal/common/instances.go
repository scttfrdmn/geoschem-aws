package common

import (
    "context"
    "fmt"
    "sort"

    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/aws"
)

// InstanceRecommendation represents a recommended instance type
type InstanceRecommendation struct {
    InstanceType    string
    VCPUs          int
    Memory         float64 // GB
    PricePerHour   float64 // USD
    Architecture   string  // x86_64 or arm64
    UseCase        string  // Description of optimal use case
    CostEfficiency float64 // Lower is better (price per vCPU)
}

// WorkloadProfile defines the characteristics of a GeosChem workload
type WorkloadProfile struct {
    GridResolution string  // "4x5", "2x2.5", "0.5x0.625"
    SpeciesCount   int     // Number of chemical species
    Duration       int     // Expected runtime in hours
    BudgetPerHour  float64 // Maximum cost per hour
    Priority       string  // "cost", "performance", "balanced"
    Architecture   string  // "x86_64", "arm64", "any"
}

// InstanceSelector handles intelligent instance type selection
type InstanceSelector struct {
    ec2Client *ec2.Client
    region    string
}

// NewInstanceSelector creates a new instance selector
func NewInstanceSelector(cfg aws.Config, region string) *InstanceSelector {
    return &InstanceSelector{
        ec2Client: ec2.NewFromConfig(cfg),
        region:    region,
    }
}

// GetRecommendations returns recommended instance types for a workload
func (is *InstanceSelector) GetRecommendations(ctx context.Context, profile WorkloadProfile) ([]InstanceRecommendation, error) {
    // Get current pricing and availability
    instances, err := is.getAvailableInstances(ctx)
    if err != nil {
        return nil, fmt.Errorf("getting available instances: %w", err)
    }

    // Filter and score instances based on workload
    recommendations := is.scoreInstances(instances, profile)
    
    // Sort by suitability score
    sort.Slice(recommendations, func(i, j int) bool {
        return is.calculateScore(recommendations[i], profile) > is.calculateScore(recommendations[j], profile)
    })

    // Return top recommendations
    maxResults := 5
    if len(recommendations) < maxResults {
        maxResults = len(recommendations)
    }
    
    return recommendations[:maxResults], nil
}

// getAvailableInstances retrieves available instance types with current pricing
func (is *InstanceSelector) getAvailableInstances(ctx context.Context) ([]InstanceRecommendation, error) {
    // For now, return static data based on research
    // In production, this would query EC2 pricing API
    instances := []InstanceRecommendation{
        // Development tier
        {
            InstanceType:    "t3.medium",
            VCPUs:          2,
            Memory:         4.0,
            PricePerHour:   0.0416,
            Architecture:   "x86_64",
            UseCase:        "Development and testing",
            CostEfficiency: 0.0208,
        },
        
        // Standard tier - x86_64
        {
            InstanceType:    "c5.xlarge",
            VCPUs:          4,
            Memory:         8.0,
            PricePerHour:   0.17,
            Architecture:   "x86_64",
            UseCase:        "Standard simulations (4x5 degree)",
            CostEfficiency: 0.0425,
        },
        {
            InstanceType:    "c5.2xlarge", 
            VCPUs:          8,
            Memory:         16.0,
            PricePerHour:   0.34,
            Architecture:   "x86_64",
            UseCase:        "High-resolution simulations (2x2.5 degree)",
            CostEfficiency: 0.0425,
        },
        
        // Standard tier - ARM64 (Graviton)
        {
            InstanceType:    "c6g.xlarge",
            VCPUs:          4,
            Memory:         8.0,
            PricePerHour:   0.136,
            Architecture:   "arm64",
            UseCase:        "Standard simulations - 20% cost savings",
            CostEfficiency: 0.034,
        },
        {
            InstanceType:    "c6g.2xlarge",
            VCPUs:          8, 
            Memory:         16.0,
            PricePerHour:   0.272,
            Architecture:   "arm64",
            UseCase:        "High-resolution simulations - 20% cost savings",
            CostEfficiency: 0.034,
        },
        
        // Memory-optimized tier - x86_64
        {
            InstanceType:    "r5.2xlarge",
            VCPUs:          8,
            Memory:         64.0,
            PricePerHour:   0.504,
            Architecture:   "x86_64", 
            UseCase:        "Memory-intensive simulations (many species)",
            CostEfficiency: 0.063,
        },
        
        // Memory-optimized tier - ARM64
        {
            InstanceType:    "r6g.2xlarge",
            VCPUs:          8,
            Memory:         64.0,
            PricePerHour:   0.403,
            Architecture:   "arm64",
            UseCase:        "Memory-intensive simulations - 20% cost savings",
            CostEfficiency: 0.050,
        },
        
        // High-performance tier
        {
            InstanceType:    "c5.4xlarge",
            VCPUs:          16,
            Memory:         32.0,
            PricePerHour:   0.68,
            Architecture:   "x86_64",
            UseCase:        "Large-scale parallel simulations",
            CostEfficiency: 0.0425,
        },
    }

    return instances, nil
}

// scoreInstances filters and scores instances based on workload profile
func (is *InstanceSelector) scoreInstances(instances []InstanceRecommendation, profile WorkloadProfile) []InstanceRecommendation {
    var filtered []InstanceRecommendation
    
    for _, instance := range instances {
        // Filter by architecture preference
        if profile.Architecture != "any" && instance.Architecture != profile.Architecture {
            continue
        }
        
        // Filter by budget
        if profile.BudgetPerHour > 0 && instance.PricePerHour > profile.BudgetPerHour {
            continue
        }
        
        // Filter by minimum requirements
        if !is.meetsMinimumRequirements(instance, profile) {
            continue
        }
        
        filtered = append(filtered, instance)
    }
    
    return filtered
}

// meetsMinimumRequirements checks if instance meets minimum workload requirements
func (is *InstanceSelector) meetsMinimumRequirements(instance InstanceRecommendation, profile WorkloadProfile) bool {
    // Minimum vCPU requirements based on grid resolution
    minVCPUs := is.getMinimumVCPUs(profile.GridResolution)
    if instance.VCPUs < minVCPUs {
        return false
    }
    
    // Minimum memory requirements based on species count
    minMemory := is.getMinimumMemory(profile.SpeciesCount, profile.GridResolution)
    if instance.Memory < minMemory {
        return false
    }
    
    return true
}

// getMinimumVCPUs returns minimum vCPUs for a given grid resolution
func (is *InstanceSelector) getMinimumVCPUs(gridResolution string) int {
    switch gridResolution {
    case "4x5":
        return 2 // Can run on 2 cores but 4 is better
    case "2x2.5":
        return 4 // Needs at least 4 cores
    case "0.5x0.625", "0.25x0.3125":
        return 8 // High-res needs more cores
    default:
        return 2 // Conservative default
    }
}

// getMinimumMemory returns minimum memory (GB) for workload characteristics
func (is *InstanceSelector) getMinimumMemory(speciesCount int, gridResolution string) float64 {
    baseMemory := 2.0 // GB base requirement
    
    // Memory scales with grid resolution
    switch gridResolution {
    case "4x5":
        baseMemory = 2.0
    case "2x2.5": 
        baseMemory = 4.0
    case "0.5x0.625":
        baseMemory = 8.0
    case "0.25x0.3125":
        baseMemory = 16.0
    }
    
    // Memory scales with number of species (roughly linear)
    speciesMemory := float64(speciesCount) * 0.02 // 20 MB per species roughly
    
    return baseMemory + speciesMemory
}

// calculateScore calculates a suitability score for an instance
func (is *InstanceSelector) calculateScore(instance InstanceRecommendation, profile WorkloadProfile) float64 {
    score := 100.0 // Base score
    
    // Priority-based scoring
    switch profile.Priority {
    case "cost":
        // Lower cost is better
        score -= instance.PricePerHour * 100
        // Bonus for ARM64 cost savings
        if instance.Architecture == "arm64" {
            score += 20
        }
        
    case "performance":
        // More vCPUs is better
        score += float64(instance.VCPUs) * 5
        // Penalize memory-optimized if not needed
        memoryRatio := instance.Memory / float64(instance.VCPUs)
        if memoryRatio > 4 && profile.SpeciesCount < 200 {
            score -= 10 // Over-provisioned memory
        }
        
    case "balanced":
    default:
        // Balance cost and performance
        score -= instance.PricePerHour * 50
        score += float64(instance.VCPUs) * 2
        if instance.Architecture == "arm64" {
            score += 10 // Moderate bonus for ARM64
        }
    }
    
    // Penalize over-provisioning
    minVCPUs := is.getMinimumVCPUs(profile.GridResolution)
    if instance.VCPUs > minVCPUs*3 {
        score -= 15 // Likely over-provisioned
    }
    
    minMemory := is.getMinimumMemory(profile.SpeciesCount, profile.GridResolution)
    if instance.Memory > minMemory*2 {
        score -= 10 // Memory over-provisioned
    }
    
    return score
}

// FormatRecommendations returns a human-readable string of recommendations
func FormatRecommendations(recommendations []InstanceRecommendation, profile WorkloadProfile) string {
    if len(recommendations) == 0 {
        return "No suitable instances found for your requirements."
    }
    
    result := fmt.Sprintf("💡 Instance Recommendations for %s simulation:\n\n", profile.GridResolution)
    
    for i, rec := range recommendations {
        rank := ""
        switch i {
        case 0:
            rank = "🥇 BEST"
        case 1:
            rank = "🥈 GOOD"
        case 2:
            rank = "🥉 OK"
        default:
            rank = fmt.Sprintf("#%d", i+1)
        }
        
        costPerDay := rec.PricePerHour * 24
        
        result += fmt.Sprintf("%s: %s\n", rank, rec.InstanceType)
        result += fmt.Sprintf("   💻 %d vCPUs, %.0f GB RAM (%s)\n", 
            rec.VCPUs, rec.Memory, rec.Architecture)
        result += fmt.Sprintf("   💰 $%.3f/hour ($%.2f/day)\n", 
            rec.PricePerHour, costPerDay)
        result += fmt.Sprintf("   📋 %s\n", rec.UseCase)
        result += "\n"
    }
    
    return result
}

// EstimateCost calculates estimated cost for a workload
func EstimateCost(instance InstanceRecommendation, durationHours int) (float64, string) {
    totalCost := instance.PricePerHour * float64(durationHours)
    
    var timeframe string
    if durationHours <= 24 {
        timeframe = fmt.Sprintf("%d hours", durationHours)
    } else {
        days := durationHours / 24
        hours := durationHours % 24
        if hours == 0 {
            timeframe = fmt.Sprintf("%d days", days)
        } else {
            timeframe = fmt.Sprintf("%d days, %d hours", days, hours)
        }
    }
    
    return totalCost, timeframe
}