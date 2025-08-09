output "iam_roles_summary" {
  description = "Summary of created IAM roles"
  value = {
    ec2_builder_role_name       = module.iam.ec2_builder_role_name
    ec2_instance_profile_name   = module.iam.ec2_builder_instance_profile_name
    batch_execution_role_arn    = module.iam.batch_execution_role_arn
    ecs_task_execution_role_arn = module.iam.ecs_task_execution_role_arn
  }
}

output "test_ecr_repository" {
  description = "Test ECR repository details"
  value = {
    name = aws_ecr_repository.test_repo.name
    arn  = aws_ecr_repository.test_repo.arn
    url  = aws_ecr_repository.test_repo.repository_url
  }
}

output "next_steps" {
  description = "Next steps for testing"
  value = <<-EOT
    
🎉 IAM Roles Created Successfully!

Created Resources:
- EC2 Builder Role: ${module.iam.ec2_builder_role_name}
- Instance Profile: ${module.iam.ec2_builder_instance_profile_name}
- Batch Execution Role: ${module.iam.batch_execution_role_arn}
- ECS Task Execution Role: ${module.iam.ecs_task_execution_role_arn}

Test the roles:
1. Check role permissions:
   aws iam get-role --role-name ${module.iam.ec2_builder_role_name} --profile ${var.aws_profile}

2. List attached policies:
   aws iam list-attached-role-policies --role-name ${module.iam.ec2_builder_role_name} --profile ${var.aws_profile}

3. Test ECR access (would be done by EC2 instance):
   aws ecr describe-repositories --repository-names ${aws_ecr_repository.test_repo.name} --profile ${var.aws_profile}

To clean up:
   terraform destroy -auto-approve
   
EOT
}