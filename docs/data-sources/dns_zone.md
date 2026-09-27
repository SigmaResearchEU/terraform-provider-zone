---
page_title: "zone_dns_zone Data Source - terraform-provider-zone"
subcategory: ""
description: |-
  Retrieves information about a DNS zone on Zone.EU.
---

# zone_dns_zone (Data Source)

Retrieves information about a DNS zone on Zone.EU.

## Example Usage

```terraform
data "zone_dns_zone" "main" {
  name = "example.com"
}

output "zone_active" {
  value = data.zone_dns_zone.main.active
}

output "zone_ipv6" {
  value = data.zone_dns_zone.main.ipv6
}
```

## Schema

### Required

- `name` (String) The DNS zone name (domain name, e.g., example.com).

### Read-Only

- `id` (String) The ID of this data source.
- `active` (Boolean) Whether the zone is active.
- `ipv6` (Boolean) Whether IPv6 is enabled for the zone.
