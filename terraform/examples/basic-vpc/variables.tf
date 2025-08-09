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

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
  default     = "geoschem-test"
}