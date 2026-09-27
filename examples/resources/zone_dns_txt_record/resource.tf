resource "zone_dns_txt_record" "spf" {
  zone        = "example.com"
  name        = "example.com"
  destination = "v=spf1 include:_spf.google.com ~all"
}
