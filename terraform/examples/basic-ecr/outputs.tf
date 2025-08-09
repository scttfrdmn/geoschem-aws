output "ecr_repository_url" {
  description = "URL of the created ECR repository"
  value       = module.ecr.repository_url
}

output "ecr_repository_arn" {
  description = "ARN of the created ECR repository"
  value       = module.ecr.repository_arn
}

output "next_steps" {
  description = "Next steps for testing"
  value = <<-EOT
    
🎉 ECR Repository Created Successfully!

Repository URL: ${module.ecr.repository_url}

Test the repository:
1. Get login token:
   aws ecr get-login-password --profile ${var.aws_profile} --region ${var.aws_region} | docker login --username AWS --password-stdin ${module.ecr.repository_url}

2. Tag and push a test image:
   docker tag hello-world:latest ${module.ecr.repository_url}:test
   docker push ${module.ecr.repository_url}:test

3. List images:
   aws ecr list-images --repository-name ${var.repository_name} --profile ${var.aws_profile} --region ${var.aws_region}

To clean up:
   terraform destroy -auto-approve
   
EOT
}