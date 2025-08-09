# AWS Quota Management for GeosChem Platform

This guide helps you understand and manage AWS service quotas for the GeosChem platform.

## Understanding AWS Quotas

AWS service quotas (formerly called service limits) are the maximum number of resources you can create in your AWS account. Some quotas are soft limits that can be increased, while others are hard limits.

## GeosChem Platform Quota Requirements

### Critical Quotas to Monitor

| Service | Quota Name | Typical Limit | Platform Need | Impact if Exceeded |
|---------|------------|---------------|---------------|-------------------|
| EC2 | Running On-Demand Instances | 20-1000 | 1-10 concurrent | Build failures |
| EC2 | Key Pairs per Region | 5000 | 1+ | Cannot launch instances |
| ECR | Repositories per Region | 10000 | 1+ | Cannot store containers |
| Batch | Compute Environments | 50 | 1+ | Cannot execute jobs |
| S3 | Buckets per Account | 100 | 1-5 | Cannot store results |

### Build Phase Quotas

During container building, you'll need:
- **1-10 EC2 instances** running simultaneously (for matrix builds)
- **Instance types**: c5.2xlarge (x86_64) or c6g.2xlarge (ARM64)
- **Estimated cost**: $0.34/hour (c5.2xlarge) or $0.27/hour (c6g.2xlarge)

### Runtime Phase Quotas

For GeosChem simulations:
- **Variable EC2 instances** via AWS Batch (depends on simulation queue)
- **Compute-optimized instances** recommended
- **Memory-optimized instances** for large datasets

## Checking Your Current Quotas

### Using the Platform Tool
```bash
# Check quotas before building
go run cmd/builder/main.go --check-quotas --profile aws --region us-east-1
```

### Using AWS CLI
```bash
# List EC2 quotas
aws service-quotas list-service-quotas \
    --service-code ec2 \
    --query 'Quotas[?contains(QuotaName,`Running On-Demand`)].[QuotaName,Value,Adjustable]' \
    --profile aws

# Check current EC2 usage
aws ec2 describe-instances \
    --filters "Name=instance-state-name,Values=running" \
    --query 'length(Reservations[].Instances[])' \
    --profile aws
```

## Requesting Quota Increases

### Automated Request (Recommended)
The platform can help automate quota increase requests:

```bash
# Future feature - automated quota requests
go run cmd/builder/main.go --request-quota-increase --service ec2 --quota-code L-1216C47A
```

### Manual Request via AWS Console
1. Go to **Service Quotas** in AWS Console
2. Search for the service (e.g., "Amazon Elastic Compute Cloud")
3. Find the specific quota
4. Click "Request quota increase"
5. Enter your desired limit and justification

### Using AWS CLI
```bash
# Request increase for Running On-Demand instances
aws service-quotas request-service-quota-increase \
    --service-code ec2 \
    --quota-code L-1216C47A \
    --desired-value 100 \
    --profile aws

# Track request status
aws service-quotas list-requested-service-quota-change-history \
    --service-code ec2 \
    --profile aws
```

## Common Quota Increase Justifications

### EC2 Instance Quota
```
Business Justification: Running atmospheric chemistry simulations using GeosChem 
requires multiple concurrent compute instances for container building and batch 
job execution. Current limit of [X] instances prevents efficient operation of 
our research workflows.

Technical Details:
- Platform: GeosChem AWS Platform on Rocky Linux 9
- Instance types: c5.2xlarge, c6g.2xlarge, m5.xlarge-m5.24xlarge
- Usage pattern: Burst usage during container builds, sustained during simulations
- Expected concurrent instances: [Y] instances
- Business impact: Research delays, inefficient resource utilization
```

### ECR Repository Quota
```
Business Justification: Scientific computing platform requires multiple container 
repositories for different compiler/MPI combinations of GeosChem atmospheric 
chemistry software.

Technical Details:
- Multiple container variants per architecture (Intel, GCC, AMD compilers)
- ARM64 and x86_64 architecture support
- Version management requires separate repositories
- Expected repositories: [Y] repositories
```

## Monitoring and Alerting

### CloudWatch Alarms
Set up alarms for quota usage:

```bash
# Create alarm for EC2 instance usage
aws cloudwatch put-metric-alarm \
    --alarm-name "EC2-Instance-Usage-High" \
    --alarm-description "EC2 instance usage approaching quota" \
    --metric-name "Usage" \
    --namespace "AWS/Usage" \
    --statistic "Maximum" \
    --dimensions Name=Type,Value=Resource Name=Resource,Value="Running On-Demand instances" \
    --period 300 \
    --evaluation-periods 2 \
    --threshold 18 \
    --comparison-operator GreaterThanThreshold \
    --profile aws
```

### Cost Considerations

#### Instance Hours During Builds
- **Single build**: ~1 hour × $0.34 = $0.34
- **Architecture matrix**: ~2 hours × $0.34 = $0.68  
- **Full matrix**: ~4 hours × $0.34 = $1.36

#### Monthly Quota Costs (if maxed out)
- **20 instances × 24h × 30 days × $0.34 = $4,896**
- Most users need much less - monitor actual usage

## Troubleshooting Common Issues

### "Request limit exceeded" Error
```bash
# Check current usage vs limits
aws service-quotas get-service-quota \
    --service-code ec2 \
    --quota-code L-1216C47A \
    --profile aws

# List all running instances
aws ec2 describe-instances \
    --filters "Name=instance-state-name,Values=running,pending" \
    --profile aws
```

### Request Denied
- **Business justification insufficient**: Provide more detail about research needs
- **Usage history**: AWS may want to see historical usage patterns
- **Gradual increases**: Start with smaller increases (e.g., 20 → 50 → 100)

### Regional Limitations  
- Some regions have lower default quotas
- Consider using multiple regions for large workloads
- **us-east-1, us-west-2, eu-west-1** typically have higher quotas

## Best Practices

1. **Monitor proactively** - Don't wait until you hit limits
2. **Request early** - Quota increases can take 24-48 hours
3. **Be specific** - Provide detailed justification and expected usage
4. **Use multiple regions** - Distribute workloads if needed
5. **Clean up resources** - Terminate unused instances to stay within limits
6. **Set up alerts** - Get notified before hitting quotas

## Future Automation

The platform will include automated quota management:

- **Pre-build checks** - Verify sufficient quota before starting builds
- **Automated requests** - Submit quota increase requests when needed
- **Usage prediction** - Estimate quota needs based on build matrix
- **Cost optimization** - Recommend most cost-effective quota levels

## Support

If you encounter quota-related issues:

1. Use `--check-quotas` flag for diagnosis
2. Check AWS Service Health Dashboard for known issues
3. Contact AWS Support for complex quota scenarios
4. Review the [AWS Service Quotas User Guide](https://docs.aws.amazon.com/servicequotas/latest/userguide/)