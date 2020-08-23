# AMI
data "aws_ami" "amazon_linux" {
  most_recent = true

  owners = ["amazon"]

  filter {
    name = "name"

    values = [
      "amzn2-ami-hvm-*-x86_64-gp2",
    ]
  }

  filter {
    name = "owner-alias"

    values = [
      "amazon",
    ]
  }
}

data "template_file" "userdata" {
  template = file("${path.module}/userdata.sh")

  vars = {
    kvdi_hostname    = "${var.kvdi_host}.${var.dns_domain}"
    acme_email       = var.acme_email
    use_lets_encrypt = var.use_lets_encrypt
  }
}

resource "aws_key_pair" "this" {
  key_name   = "${var.name}-key"
  public_key = local.public_key
}

resource "aws_eip" "this" {
  vpc      = true
  instance = module.ec2.id[0]
}

module "ec2" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "2.15.0"

  instance_count = 1

  name          = var.name
  key_name      = aws_key_pair.this.key_name
  ami           = data.aws_ami.amazon_linux.id
  instance_type = var.instance_class
  subnet_id     = module.vpc.public_subnets[0]

  vpc_security_group_ids      = [module.sg.this_security_group_id]
  associate_public_ip_address = true

  user_data_base64 = base64encode(data.template_file.userdata.rendered)

  root_block_device = [
    {
      volume_type = "gp2"
      volume_size = 60
    },
  ]

  tags = {
    "Name" = var.name
  }
}
