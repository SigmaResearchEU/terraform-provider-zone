resource "zone_dns_mx_record" "mail" {
  zone        = "example.com"
  name        = "example.com"
  destination = "mail.example.com"
  priority    = 10
}
