resource "zone_dns_url_record" "redirect" {
  zone          = "example.com"
  name          = "old.example.com"
  destination   = "https://new.example.com/page"
  redirect_type = 301
}
