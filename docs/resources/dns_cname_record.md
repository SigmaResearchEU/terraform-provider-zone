---
page_title: "zone_dns_cname_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS CNAME record on Zone.EU.
---

# zone_dns_cname_record (Resource)

Manages a DNS CNAME record on Zone.EU. CNAME records create an alias from one hostname to another.

## Example Usage

```terraform
resource "zone_dns_cname_record" "blog" {
  zone        = "example.com"
  name        = "blog.example.com"
  destination = "bloghost.example.net"
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The hostname for the CNAME record (FQDN, e.g., blog.example.com).
- `destination` (String) The canonical hostname this record points to.

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_cname_record.blog example.com/123456
```
