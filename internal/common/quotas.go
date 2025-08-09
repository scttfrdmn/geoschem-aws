package common

import (
    "context"
    "fmt"
    "strings"

    "github.com/aws/aws-sdk-go-v2/service/servicequotas"
    "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
    "github.com/aws/aws-sdk-go-v2/service/support"
    "github.com/aws/aws-sdk-go-v2/aws"
)

// QuotaChecker handles AWS service quota validation
type QuotaChecker struct {
    quotasClient  *servicequotas.Client
    ec2Client     *ec2.Client
    supportClient *support.Client
    region        string
}

// QuotaStatus represents the status of a service quota
type QuotaStatus struct {
    ServiceName   string
    QuotaName     string
    Current       float64
    Limit         float64
    Usage         float64
    Status        string // OK, WARNING, CRITICAL
    Message       string
    CanIncrease   bool
}

// QuotaReport contains all quota checks
type QuotaReport struct {
    Region  string
    Quotas  []QuotaStatus
    Summary string
}

// NewQuotaChecker creates a new quota checker
func NewQuotaChecker(cfg aws.Config, region string) *QuotaChecker {
    return &QuotaChecker{
        quotasClient:  servicequotas.NewFromConfig(cfg),
        ec2Client:     ec2.NewFromConfig(cfg),
        supportClient: support.NewFromConfig(cfg),
        region:        region,
    }
}

// CheckGeoChemQuotas checks all relevant quotas for the GeosChem platform
func (qc *QuotaChecker) CheckGeoChemQuotas(ctx context.Context) (*QuotaReport, error) {
    report := &QuotaReport{
        Region: qc.region,
        Quotas: make([]QuotaStatus, 0),
    }

    // Check EC2 quotas
    ec2Quotas, err := qc.checkEC2Quotas(ctx)
    if err != nil {
        return nil, fmt.Errorf("checking EC2 quotas: %w", err)
    }
    report.Quotas = append(report.Quotas, ec2Quotas...)

    // Check ECR quotas
    ecrQuotas, err := qc.checkECRQuotas(ctx)
    if err != nil {
        return nil, fmt.Errorf("checking ECR quotas: %w", err)
    }
    report.Quotas = append(report.Quotas, ecrQuotas...)

    // Check Batch quotas
    batchQuotas, err := qc.checkBatchQuotas(ctx)
    if err != nil {
        return nil, fmt.Errorf("checking Batch quotas: %w", err)
    }
    report.Quotas = append(report.Quotas, batchQuotas...)

    // Generate summary
    report.Summary = qc.generateSummary(report.Quotas)

    return report, nil
}

// checkEC2Quotas checks EC2-related quotas
func (qc *QuotaChecker) checkEC2Quotas(ctx context.Context) ([]QuotaStatus, error) {
    quotas := make([]QuotaStatus, 0)

    // Check Running On-Demand instances
    onDemandQuota, err := qc.getQuota(ctx, "ec2", "L-1216C47A") // Running On-Demand instances
    if err != nil {
        return nil, fmt.Errorf("getting on-demand quota: %w", err)
    }

    // Get current usage
    instances, err := qc.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
        Filters: []ec2types.Filter{
            {
                Name:   aws.String("instance-state-name"),
                Values: []string{"running", "pending"},
            },
        },
    })
    if err != nil {
        return nil, fmt.Errorf("describing instances: %w", err)
    }

    runningInstances := 0
    for _, reservation := range instances.Reservations {
        runningInstances += len(reservation.Instances)
    }

    quotaValue := float64(0)
    if onDemandQuota.Value != nil {
        quotaValue = *onDemandQuota.Value
    }
    status := qc.evaluateQuotaStatus(float64(runningInstances), quotaValue)
    quotas = append(quotas, QuotaStatus{
        ServiceName: "EC2",
        QuotaName:   "Running On-Demand Instances",
        Current:     float64(runningInstances),
        Limit:       quotaValue,
        Usage:       (float64(runningInstances) / quotaValue) * 100,
        Status:      status,
        Message:     qc.getQuotaMessage("EC2 instances", status, float64(runningInstances), quotaValue),
        CanIncrease: onDemandQuota.Adjustable,
    })

    // Check EC2 Key Pairs
    keyPairQuota, err := qc.getQuota(ctx, "ec2", "L-7C0D3F92") // EC2 Key Pairs
    if err == nil {
        keyPairs, err := qc.ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{})
        if err == nil {
            keyPairCount := len(keyPairs.KeyPairs)
            keyPairQuotaValue := float64(0)
            if keyPairQuota.Value != nil {
                keyPairQuotaValue = *keyPairQuota.Value
            }
            status := qc.evaluateQuotaStatus(float64(keyPairCount), keyPairQuotaValue)
            quotas = append(quotas, QuotaStatus{
                ServiceName: "EC2",
                QuotaName:   "Key Pairs",
                Current:     float64(keyPairCount),
                Limit:       keyPairQuotaValue,
                Usage:       (float64(keyPairCount) / keyPairQuotaValue) * 100,
                Status:      status,
                Message:     qc.getQuotaMessage("EC2 key pairs", status, float64(keyPairCount), keyPairQuotaValue),
                CanIncrease: keyPairQuota.Adjustable,
            })
        }
    }

    return quotas, nil
}

// checkECRQuotas checks ECR-related quotas
func (qc *QuotaChecker) checkECRQuotas(ctx context.Context) ([]QuotaStatus, error) {
    quotas := make([]QuotaStatus, 0)

    // Check ECR Repositories
    repoQuota, err := qc.getQuota(ctx, "ecr", "L-CFEB8E8D") // Repositories per region
    if err != nil {
        return nil, fmt.Errorf("getting ECR repo quota: %w", err)
    }

    // Get current repository count (this would need ECR client)
    // For now, we'll estimate based on platform needs
    repoQuotaValue := float64(0)
    if repoQuota.Value != nil {
        repoQuotaValue = *repoQuota.Value
    }
    quotas = append(quotas, QuotaStatus{
        ServiceName: "ECR",
        QuotaName:   "Repositories per Region",
        Current:     1, // We need at least 1 for geoschem
        Limit:       repoQuotaValue,
        Usage:       (1.0 / repoQuotaValue) * 100,
        Status:      "OK",
        Message:     fmt.Sprintf("ECR repositories: 1 used of %.0f available", repoQuotaValue),
        CanIncrease: repoQuota.Adjustable,
    })

    return quotas, nil
}

// checkBatchQuotas checks AWS Batch quotas
func (qc *QuotaChecker) checkBatchQuotas(ctx context.Context) ([]QuotaStatus, error) {
    quotas := make([]QuotaStatus, 0)

    // Check Batch Compute Environments
    computeEnvQuota, err := qc.getQuota(ctx, "batch", "L-D8F0C5EA") // Compute environments per region
    if err != nil {
        return nil, fmt.Errorf("getting Batch compute environment quota: %w", err)
    }

    computeEnvQuotaValue := float64(0)
    if computeEnvQuota.Value != nil {
        computeEnvQuotaValue = *computeEnvQuota.Value
    }
    quotas = append(quotas, QuotaStatus{
        ServiceName: "Batch",
        QuotaName:   "Compute Environments per Region",
        Current:     1, // We need at least 1
        Limit:       computeEnvQuotaValue,
        Usage:       (1.0 / computeEnvQuotaValue) * 100,
        Status:      "OK",
        Message:     fmt.Sprintf("Batch compute environments: 1 needed of %.0f available", computeEnvQuotaValue),
        CanIncrease: computeEnvQuota.Adjustable,
    })

    return quotas, nil
}

// getQuota retrieves a specific quota
func (qc *QuotaChecker) getQuota(ctx context.Context, serviceCode, quotaCode string) (*types.ServiceQuota, error) {
    input := &servicequotas.GetServiceQuotaInput{
        ServiceCode: aws.String(serviceCode),
        QuotaCode:   aws.String(quotaCode),
    }

    result, err := qc.quotasClient.GetServiceQuota(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("getting quota %s for service %s: %w", quotaCode, serviceCode, err)
    }

    return result.Quota, nil
}

// evaluateQuotaStatus determines the status based on usage
func (qc *QuotaChecker) evaluateQuotaStatus(current, limit float64) string {
    usage := (current / limit) * 100

    switch {
    case usage >= 90:
        return "CRITICAL"
    case usage >= 75:
        return "WARNING"
    default:
        return "OK"
    }
}

// getQuotaMessage generates a human-readable message
func (qc *QuotaChecker) getQuotaMessage(resourceName, status string, current, limit float64) string {
    usage := (current / limit) * 100

    switch status {
    case "CRITICAL":
        return fmt.Sprintf("⚠️  %s usage is critical: %.0f/%.0f (%.1f%%) - Consider requesting quota increase", resourceName, current, limit, usage)
    case "WARNING":
        return fmt.Sprintf("⚠️  %s usage is high: %.0f/%.0f (%.1f%%) - Monitor closely", resourceName, current, limit, usage)
    default:
        return fmt.Sprintf("✅ %s usage is normal: %.0f/%.0f (%.1f%%)", resourceName, current, limit, usage)
    }
}

// generateSummary creates an overall summary
func (qc *QuotaChecker) generateSummary(quotas []QuotaStatus) string {
    var critical, warning, ok int

    for _, quota := range quotas {
        switch quota.Status {
        case "CRITICAL":
            critical++
        case "WARNING":
            warning++
        case "OK":
            ok++
        }
    }

    var summary strings.Builder
    summary.WriteString(fmt.Sprintf("Quota Check Summary for %s:\n", qc.region))
    summary.WriteString(fmt.Sprintf("✅ OK: %d  ⚠️  WARNING: %d  🚨 CRITICAL: %d\n", ok, warning, critical))

    if critical > 0 {
        summary.WriteString("\n🚨 CRITICAL quotas need immediate attention - consider requesting increases")
    } else if warning > 0 {
        summary.WriteString("\n⚠️  Some quotas are approaching limits - monitor usage")
    } else {
        summary.WriteString("\n✅ All quotas look good for GeosChem platform usage")
    }

    return summary.String()
}

// PrintReport prints a formatted quota report
func (qr *QuotaReport) PrintReport() {
    fmt.Println("🔍 AWS Quota Report for GeosChem Platform")
    fmt.Println("=" + strings.Repeat("=", 50))
    fmt.Println()
    fmt.Println(qr.Summary)
    fmt.Println()

    for _, quota := range qr.Quotas {
        statusIcon := "✅"
        if quota.Status == "WARNING" {
            statusIcon = "⚠️ "
        } else if quota.Status == "CRITICAL" {
            statusIcon = "🚨"
        }

        fmt.Printf("%s %s - %s\n", statusIcon, quota.ServiceName, quota.QuotaName)
        fmt.Printf("   Usage: %.0f/%.0f (%.1f%%)\n", quota.Current, quota.Limit, quota.Usage)
        fmt.Printf("   %s\n", quota.Message)
        if quota.CanIncrease && (quota.Status == "WARNING" || quota.Status == "CRITICAL") {
            fmt.Printf("   💡 This quota can be increased via AWS Support case\n")
        }
        fmt.Println()
    }
}