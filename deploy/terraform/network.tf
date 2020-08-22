# VPC
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "2.48.0"

  name = "${var.name}-vpc"

  cidr = local.vpc_cidr
  azs  = local.azs
  public_subnets = [for i in range(length(local.azs)) :
    cidrsubnet(local.vpc_cidr, 8, i + 1)
  ]

  vpc_tags = {
    Name = "${var.name}-vpc"
  }
}
