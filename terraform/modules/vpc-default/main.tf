terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Use default VPC - perfect for HPC workloads (simple, no extra costs)
data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# Get first subnet for consistent deployments
data "aws_subnet" "default" {
  id = data.aws_subnets.default.ids[0]
}

# Security Group for HPC compute instances
resource "aws_security_group" "hpc_compute" {
  name_prefix = "${var.name_prefix}-hpc-compute"
  vpc_id      = data.aws_vpc.default.id
  description = "Security group for HPC compute instances"

  # SSH access
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # All outbound traffic
  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-hpc-compute-sg"
  })
}