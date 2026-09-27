resource "zone_dns_ns_record" "subdomain" {
  zone        = "example.com"
  name        = "subdomain.example.com"
  destination = "ns1.otherdns.com"
}
