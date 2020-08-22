variable "ext_ip" {
  type        = string
  description = "The external IP to allow access to kVDI."
}

variable "name" {
  type = string
  description = "The name prefix to use for resources."
  default = "kvdi-demo"
}

variable "instance_class" {
  type        = string
  description = "The instance class to use for the kVDI demo server."
  default     = "t3.large"
}

variable "region" {
  type        = string
  description = "The AWS region to deploy resources in."
  default     = "eu-central-1"
}

variable "vpc_cidr" {
  type        = string
  description = "The CIDR block to use for the kvdi demo VPC."
  default     = "10.0.0.0/16"
}

variable "public_key" {
  type        = string
  description = "The public key to authorize for SSH access to the instance. Defaults to `~/.ssh/id_rsa.pub`."
  default     = ""
}