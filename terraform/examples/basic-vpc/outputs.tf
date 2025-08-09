output "vpc_summary" {
  description = "Summary of created VPC resources"
  value = {
    vpc_id               = module.vpc.vpc_id
    vpc_cidr            = module.vpc.vpc_cidr_block
    public_subnet_id    = module.vpc.public_subnet_id
    availability_zone   = module.vpc.availability_zone
  }
}

output "security_groups" {
  description = "Security group information"
  value = {
    hpc_compute_sg_id = module.vpc.hpc_compute_security_group_id
  }
}

output "next_steps" {
  description = "Next steps for testing"
  value = <<-EOT
    
🌐 Default VPC Setup Complete!

Network Resources:
- VPC: ${module.vpc.vpc_id} (${module.vpc.vpc_cidr_block})
- Public Subnet: ${module.vpc.public_subnet_id}
- Availability Zone: ${module.vpc.availability_zone}
- All Subnets: ${length(module.vpc.public_subnet_ids)} available

Security Groups:
- HPC Compute: ${module.vpc.hpc_compute_security_group_id}

Perfect for HPC/Open Science:
✓ Uses existing default VPC (no extra resources)
✓ Public subnets (direct internet access)
✓ No NAT Gateway fees ($45/month saved!)
✓ Simple, cost-effective design
✓ Ready for compute workloads

Test the setup:
1. List VPC details:
   aws ec2 describe-vpcs --vpc-ids ${module.vpc.vpc_id} --profile ${var.aws_profile}

2. Test subnet:
   aws ec2 describe-subnets --subnet-ids ${module.vpc.public_subnet_id} --profile ${var.aws_profile}

To clean up (removes only security group):
   terraform destroy -auto-approve
   
EOT
}