
data "aws_route53_zone" "this" {
  count = var.use_lets_encrypt ? 1 : 0

  name         = "${var.dns_domain}."
  private_zone = false
}

resource "aws_route53_record" "www" {
  count = var.use_lets_encrypt ? 1 : 0

  zone_id = data.aws_route53_zone.this[0].zone_id
  name    = "${var.kvdi_host}.${var.dns_domain}"
  type    = "A"
  ttl     = "300"
  records = [aws_eip.this.public_ip]
}