---
page_title: "zone_dns_aaaa_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS AAAA record on Zone.EU.
---

# zone_dns_aaaa_record (Resource)

Manages a DNS AAAA record on Zone.EU. AAAA records map hostnames to IPv6 addresses.

## Example Usage

```terraform
resource "zone_dns_aaaa_record" "www" {
  zone        = "example.com"
  name        = "www.example.com"
  destination = "2001:db8::1"
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The hostname for the AAAA record (FQDN, e.g., www.example.com).
- `destination` (String) The IPv6 address the record points to. Must be a valid IPv6 address.

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_aaaa_record.www example.com/123456
```
