---
page_title: "zone_dns_ns_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS NS record on Zone.EU.
---

# zone_dns_ns_record (Resource)

Manages a DNS NS record on Zone.EU. NS records delegate a DNS zone to use the given authoritative name servers.

## Example Usage

```terraform
resource "zone_dns_ns_record" "subdomain" {
  zone        = "example.com"
  name        = "subdomain.example.com"
  destination = "ns1.otherdns.com"
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The hostname for the NS record (FQDN, e.g., subdomain.example.com).
- `destination` (String) The nameserver hostname.

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_ns_record.subdomain example.com/123456
```
