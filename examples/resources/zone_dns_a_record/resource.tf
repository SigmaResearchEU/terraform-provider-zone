resource "zone_dns_a_record" "www" {
  zone        = "example.com"
  name        = "www.example.com"
  destination = "192.168.1.1"
}
