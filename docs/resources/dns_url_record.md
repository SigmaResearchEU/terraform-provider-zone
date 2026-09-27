---
page_title: "zone_dns_url_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS URL redirect record on Zone.EU.
---

# zone_dns_url_record (Resource)

Manages a DNS URL redirect record on Zone.EU. This is a Zone.EU-specific record type that provides URL redirection at the DNS level.

## Example Usage

```terraform
resource "zone_dns_url_record" "redirect" {
  zone          = "example.com"
  name          = "old.example.com"
  destination   = "https://new.example.com/page"
  redirect_type = 301
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The hostname to redirect from (FQDN, e.g., old.example.com).
- `destination` (String) The URL to redirect to. Must be a valid URL starting with `http://` or `https://`.
- `redirect_type` (Number) The HTTP redirect status code. Must be 301 or 302:
  - 301: Permanent redirect
  - 302: Temporary redirect

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_url_record.redirect example.com/123456
```
