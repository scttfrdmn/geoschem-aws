output "repository_url" {
  description = "URL of the ECR repository"
  value       = aws_ecr_repository.geoschem.repository_url
}

output "repository_arn" {
  description = "ARN of the ECR repository"
  value       = aws_ecr_repository.geoschem.arn
}

output "repository_name" {
  description = "Name of the ECR repository"
  value       = aws_ecr_repository.geoschem.name
}

output "registry_id" {
  description = "Registry ID of the ECR repository"
  value       = aws_ecr_repository.geoschem.registry_id
}