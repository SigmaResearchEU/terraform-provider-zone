data "zone_dns_zone" "main" {
  name = "example.com"
}

output "zone_active" {
  value = data.zone_dns_zone.main.active
}

output "zone_ipv6" {
  value = data.zone_dns_zone.main.ipv6
}
