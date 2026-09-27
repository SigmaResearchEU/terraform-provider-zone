resource "zone_dns_caa_record" "letsencrypt" {
  zone        = "example.com"
  name        = "example.com"
  destination = "letsencrypt.org"
  flag        = 0
  tag         = "issue"
}
