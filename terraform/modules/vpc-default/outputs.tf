output "vpc_id" {
  description = "ID of the default VPC"
  value       = data.aws_vpc.default.id
}

output "vpc_cidr_block" {
  description = "CIDR block of the default VPC"
  value       = data.aws_vpc.default.cidr_block
}

output "public_subnet_id" {
  description = "ID of the first default subnet"
  value       = data.aws_subnet.default.id
}

output "public_subnet_ids" {
  description = "IDs of all default subnets"
  value       = data.aws_subnets.default.ids
}

output "availability_zone" {
  description = "Availability zone of first subnet"
  value       = data.aws_subnet.default.availability_zone
}

output "hpc_compute_security_group_id" {
  description = "ID of the HPC compute security group"
  value       = aws_security_group.hpc_compute.id
}