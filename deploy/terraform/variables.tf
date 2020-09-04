variable "ext_ip" {
  type        = string
  description = "The external IP to allow access to kVDI."
}

variable "name" {
  type        = string
  description = "The name prefix to use for resources."
  default     = "kvdi-demo"
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

variable "use_lets_encrypt" {
  type        = bool
  description = "Whether to use Let's Encrypt to generate signed server certificates. When set to true, DNS records will be created as well."
  default     = true
}

variable "dns_domain" {
  type        = string
  description = "The DNS domain to create a record in."
  default     = "kvdi.io"
}

variable "kvdi_host" {
  type        = string
  description = "The top-level host to use in the DNS name for the instance."
  default     = "demo"
}

variable "acme_email" {
  type        = string
  description = "The email address to use for ACME registration."
  default     = "38474291+tinyzimmer@users.noreply.github.com"
}

variable "prom_operator_version" {
  type        = string
  description = "The version of prometheus-operator to install alongside kVDI"
  default     = "v0.41.0"
}