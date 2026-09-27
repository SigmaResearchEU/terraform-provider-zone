---
page_title: "zone_dns_txt_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS TXT record on Zone.EU.
---

# zone_dns_txt_record (Resource)

Manages a DNS TXT record on Zone.EU. TXT records are used for various purposes including SPF, DKIM, and domain verification.

## Example Usage

```terraform
resource "zone_dns_txt_record" "spf" {
  zone        = "example.com"
  name        = "example.com"
  destination = "v=spf1 include:_spf.google.com ~all"
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The hostname for the TXT record (FQDN, e.g., example.com).
- `destination` (String) The text content of the record.

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_txt_record.spf example.com/123456
```
