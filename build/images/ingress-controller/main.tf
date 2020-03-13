terraform {
  backend "local" {
    path = "terraform.tfstate"
  }
}

provider "aws" {
  access_key = var.access_key
  secret_key = var.secret_key
  region     = var.region
}

resource "aws_vpc" "vpc" {
  cidr_block           = var.cidr_vpc
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = {
    "Project" = var.project_tag
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    "Project" = var.project_tag
  }
}

resource "aws_subnet" "subnet_public" {
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = var.cidr_subnet
  map_public_ip_on_launch = "true"
  availability_zone       = var.availability_zone
  tags = {
    "Project" = var.project_tag
  }
}

resource "aws_route_table" "rtb_public" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    "Project" = var.project_tag
  }
}

resource "aws_route_table_association" "rta_subnet_public" {
  subnet_id      = aws_subnet.subnet_public.id
  route_table_id = aws_route_table.rtb_public.id
}

resource "aws_security_group" "allow_ssh" {
  name   = "ssh"
  vpc_id = aws_vpc.vpc.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    "Project" = var.project_tag
  }
}

resource "tls_private_key" "bootstrap_private_key" {
    algorithm = "RSA"
    rsa_bits  = "4096"
}

resource "aws_key_pair" "ssh_key" {
  key_name   = "ssh-key_${var.project_tag}"
  public_key = chomp(tls_private_key.bootstrap_private_key.public_key_openssh)
}

resource "local_file" "public_key_openssh" {
  count      = 1
  depends_on = [tls_private_key.bootstrap_private_key]
  content    = tls_private_key.bootstrap_private_key.public_key_pem
  filename   = "id_rsa.pub"
}

resource "local_file" "private_key_openssh" {
  count      = 1
  depends_on = [tls_private_key.bootstrap_private_key]
  content    = tls_private_key.bootstrap_private_key.private_key_pem
  filename   = "id_rsa"
}

data "aws_ami" "latest-ubuntu" {
  most_recent = true

  owners = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name = "block-device-mapping.volume-type"
    values = ["gp2"]
  }
}

resource "aws_spot_instance_request" "build_worker" {
  ami                    = data.aws_ami.latest-ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.subnet_public.id
  vpc_security_group_ids = [aws_security_group.allow_ssh.id]

  valid_until = var.valid_until

  key_name = aws_key_pair.ssh_key.key_name

  spot_price = "2"
  spot_type = "one-time"

  ebs_optimized = true

  root_block_device {
    volume_size           = 32
    volume_type           = "gp2"
    delete_on_termination = true
  }

  wait_for_fulfillment = true
  instance_initiated_shutdown_behavior = "terminate"

  associate_public_ip_address = true

  tags = {
    "Project" = var.project_tag
  }

  connection {
    host        = coalesce(self.public_ip, self.private_ip)
    type        = "ssh"
    user        = "ubuntu"
    private_key = tls_private_key.bootstrap_private_key.private_key_pem
  }

  provisioner "file" {
    source      = "build-ingress-controller.sh"
    destination = "/tmp/build-ingress-controller.sh"
  }

  provisioner "file" {
    source      = "/root/env.tfvars"
    destination = "/tmp/env"
  }

  provisioner "remote-exec" {
    inline = [
      "echo Building ingress controller images...",
      "chmod +x /tmp/build-ingress-controller.sh",
      "sudo /tmp/build-ingress-controller.sh",
    ]
  }
}
