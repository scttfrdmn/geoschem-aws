terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# IAM role for EC2 instances to build containers
resource "aws_iam_role" "ec2_builder_role" {
  name = "${var.name_prefix}-ec2-builder-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# Policy for ECR access (push/pull container images)
resource "aws_iam_policy" "ecr_access_policy" {
  name        = "${var.name_prefix}-ecr-access-policy"
  description = "Policy for accessing ECR repositories"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:BatchDeleteImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:PutImage"
        ]
        Resource = var.ecr_repository_arns
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken"
        ]
        Resource = "*"
      }
    ]
  })

  tags = var.tags
}

# Policy for S3 access (for build artifacts and results)
resource "aws_iam_policy" "s3_access_policy" {
  count = length(var.s3_bucket_arns) > 0 ? 1 : 0
  
  name        = "${var.name_prefix}-s3-access-policy"
  description = "Policy for accessing S3 buckets"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = flatten([
          for bucket_arn in var.s3_bucket_arns : [
            bucket_arn,
            "${bucket_arn}/*"
          ]
        ])
      }
    ]
  })

  tags = var.tags
}

# Attach policies to the EC2 builder role
resource "aws_iam_role_policy_attachment" "ec2_builder_ecr_access" {
  role       = aws_iam_role.ec2_builder_role.name
  policy_arn = aws_iam_policy.ecr_access_policy.arn
}

resource "aws_iam_role_policy_attachment" "ec2_builder_s3_access" {
  count = length(var.s3_bucket_arns) > 0 ? 1 : 0
  
  role       = aws_iam_role.ec2_builder_role.name
  policy_arn = aws_iam_policy.s3_access_policy[0].arn
}

# Instance profile for EC2 instances
resource "aws_iam_instance_profile" "ec2_builder_profile" {
  name = "${var.name_prefix}-ec2-builder-profile"
  role = aws_iam_role.ec2_builder_role.name

  tags = var.tags
}

# IAM role for AWS Batch execution
resource "aws_iam_role" "batch_execution_role" {
  name = "${var.name_prefix}-batch-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "batch.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# Attach AWS managed policy for Batch service execution
resource "aws_iam_role_policy_attachment" "batch_execution_role_policy" {
  role       = aws_iam_role.batch_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBatchServiceRole"
}

# IAM role for ECS task execution (used by Batch)
resource "aws_iam_role" "ecs_task_execution_role" {
  name = "${var.name_prefix}-ecs-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# Attach AWS managed policy for ECS task execution
resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}