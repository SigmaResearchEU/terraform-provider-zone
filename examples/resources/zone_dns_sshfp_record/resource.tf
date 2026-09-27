resource "zone_dns_sshfp_record" "server" {
  zone        = "example.com"
  name        = "server.example.com"
  destination = "abc123def456789..."
  algorithm   = 4
  type        = 2
}
