provider "aws" {
  region  = local.region
  version = "~> 3.3.0"
}

locals {
  region   = var.region
  azs      = ["${var.region}a"]
  vpc_cidr = var.vpc_cidr

  public_key = var.public_key != "" ? var.public_key : file(pathexpand("~/.ssh/id_rsa.pub"))
}

output "public_ip" {
  value = aws_eip.this.public_ip
}