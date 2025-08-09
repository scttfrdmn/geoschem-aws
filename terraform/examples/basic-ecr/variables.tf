variable "aws_profile" {
  description = "AWS profile to use"
  type        = string
  default     = "aws"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "repository_name" {
  description = "Name of the ECR repository"
  type        = string
  default     = "geoschem-test"
}