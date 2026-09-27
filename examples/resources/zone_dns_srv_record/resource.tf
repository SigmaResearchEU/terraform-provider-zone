resource "zone_dns_srv_record" "sip" {
  zone        = "example.com"
  name        = "_sip._tcp.example.com"
  destination = "sipserver.example.com"
  priority    = 10
  weight      = 5
  port        = 5060
}
