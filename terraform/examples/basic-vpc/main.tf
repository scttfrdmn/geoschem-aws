terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  profile = var.aws_profile
  region  = var.aws_region

  default_tags {
    tags = {
      Project     = "geoschem-aws"
      Environment = "development"
      Terraform   = "true"
    }
  }
}

# Use default VPC - perfect for HPC/open science (simple, cost-effective)
module "vpc" {
  source = "../../modules/vpc-default"

  name_prefix = var.name_prefix

  tags = {
    Project     = "geoschem-aws"
    Environment = "development"
    Terraform   = "true"
    TestStep    = "basic-vpc"
  }
}