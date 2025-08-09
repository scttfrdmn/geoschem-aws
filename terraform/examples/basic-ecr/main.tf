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

data "aws_caller_identity" "current" {}

module "ecr" {
  source = "../../modules/ecr"

  repository_name = var.repository_name
  allowed_principals = [
    data.aws_caller_identity.current.arn
  ]

  tags = {
    Project     = "geoschem-aws-platform"
    Environment = "development"
    Terraform   = "true"
    TestStep    = "basic-ecr"
  }
}