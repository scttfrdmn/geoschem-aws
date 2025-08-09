variable "name_prefix" {
  description = "Prefix for all IAM resource names"
  type        = string
  default     = "geoschem"
}

variable "ecr_repository_arns" {
  description = "List of ECR repository ARNs that EC2 instances need access to"
  type        = list(string)
  default     = []
}

variable "s3_bucket_arns" {
  description = "List of S3 bucket ARNs that EC2 instances need access to"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags to apply to all IAM resources"
  type        = map(string)
  default = {
    Project   = "geoschem-aws-platform"
    Terraform = "true"
  }
}