variable "repository_name" {
  description = "Name of the ECR repository"
  type        = string
  default     = "geoschem"
}

variable "allowed_principals" {
  description = "List of AWS principals allowed to access the repository"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags to apply to the ECR repository"
  type        = map(string)
  default = {
    Project   = "geoschem-aws-platform"
    Terraform = "true"
  }
}