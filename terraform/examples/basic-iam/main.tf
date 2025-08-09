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
      Project     = "geoschem-aws-platform"
      Environment = "development"
      Terraform   = "true"
    }
  }
}

# Create a test ECR repository to reference in IAM policies
resource "aws_ecr_repository" "test_repo" {
  name = "geoschem-iam-test"
  
  tags = {
    Purpose = "iam-testing"
  }
}

# Test the IAM module
module "iam" {
  source = "../../modules/iam"

  name_prefix = var.name_prefix
  
  ecr_repository_arns = [
    aws_ecr_repository.test_repo.arn
  ]
  
  # No S3 buckets for basic test
  s3_bucket_arns = []

  tags = {
    Project     = "geoschem-aws-platform"
    Environment = "development"
    Terraform   = "true"
    TestStep    = "basic-iam"
  }
}