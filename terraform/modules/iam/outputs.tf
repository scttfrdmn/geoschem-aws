output "ec2_builder_role_arn" {
  description = "ARN of the EC2 builder IAM role"
  value       = aws_iam_role.ec2_builder_role.arn
}

output "ec2_builder_role_name" {
  description = "Name of the EC2 builder IAM role"
  value       = aws_iam_role.ec2_builder_role.name
}

output "ec2_builder_instance_profile_name" {
  description = "Name of the EC2 builder instance profile"
  value       = aws_iam_instance_profile.ec2_builder_profile.name
}

output "ec2_builder_instance_profile_arn" {
  description = "ARN of the EC2 builder instance profile"
  value       = aws_iam_instance_profile.ec2_builder_profile.arn
}

output "batch_execution_role_arn" {
  description = "ARN of the Batch execution IAM role"
  value       = aws_iam_role.batch_execution_role.arn
}

output "ecs_task_execution_role_arn" {
  description = "ARN of the ECS task execution IAM role"
  value       = aws_iam_role.ecs_task_execution_role.arn
}

output "ecr_access_policy_arn" {
  description = "ARN of the ECR access policy"
  value       = aws_iam_policy.ecr_access_policy.arn
}

output "s3_access_policy_arn" {
  description = "ARN of the S3 access policy"
  value       = length(var.s3_bucket_arns) > 0 ? aws_iam_policy.s3_access_policy[0].arn : null
}