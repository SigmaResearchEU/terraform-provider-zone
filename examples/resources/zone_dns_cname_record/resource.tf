resource "zone_dns_cname_record" "blog" {
  zone        = "example.com"
  name        = "blog.example.com"
  destination = "bloghost.example.net"
}
