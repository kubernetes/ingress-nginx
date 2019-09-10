variable "access_key" {
}

variable "secret_key" {
}

variable "valid_until" {
}

variable "docker_username" {
}

variable "docker_password" {
}

variable "region" {
  default = "us-west-2"
}

variable "cidr_vpc" {
  description = "CIDR block for the VPC"
  default     = "10.4.0.0/16"
}

variable "cidr_subnet" {
  description = "CIDR block for the subnet"
  default     = "10.4.0.0/24"
}

variable "availability_zone" {
  description = "availability zone to create subnet"
  default     = "us-west-2b"
}

variable "ssh_key_path" {
  description = "Path to the SSH key"
  default     = "~/.ssh/id_rsa"
}

variable "ssh_public_key_path" {
  description = "Path to the public SSH key"
  default     = "~/.ssh/id_rsa.pub"
}

variable "instance_type" {
  description = "EC2 instance"
  default     = "c5.18xlarge"
}

variable "project_tag" {
  default = "kubernetes/ingress-nginx"
}
