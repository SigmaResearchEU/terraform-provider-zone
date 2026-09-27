resource "zone_dns_tlsa_record" "https" {
  zone              = "example.com"
  name              = "_443._tcp.example.com"
  destination       = "abc123def456789..."
  certificate_usage = 3
  selector          = 1
  matching_type     = 1
}
