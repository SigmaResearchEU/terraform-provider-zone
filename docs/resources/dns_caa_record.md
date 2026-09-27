---
page_title: "zone_dns_caa_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS CAA record on Zone.EU.
---

# zone_dns_caa_record (Resource)

Manages a DNS CAA record on Zone.EU. CAA records specify which Certificate Authorities (CAs) are allowed to issue certificates for a domain.

## Example Usage

```terraform
resource "zone_dns_caa_record" "letsencrypt" {
  zone        = "example.com"
  name        = "example.com"
  destination = "letsencrypt.org"
  flag        = 0
  tag         = "issue"
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The hostname for the CAA record (FQDN, e.g., example.com).
- `destination` (String) The value associated with the tag (e.g., CA domain).
- `flag` (Number) The CAA record flag. Must be between 0 and 255. Commonly 0 for non-critical or 128 for critical.
- `tag` (String) The CAA tag. Must be one of: `issue`, `issuewild`, or `iodef`.

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_caa_record.letsencrypt example.com/123456
```
