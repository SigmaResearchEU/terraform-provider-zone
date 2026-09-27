---
page_title: "zone_dns_mx_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS MX record on Zone.EU.
---

# zone_dns_mx_record (Resource)

Manages a DNS MX record on Zone.EU. MX records specify mail servers responsible for accepting email on behalf of a domain.

## Example Usage

```terraform
resource "zone_dns_mx_record" "mail" {
  zone        = "example.com"
  name        = "example.com"
  destination = "mail.example.com"
  priority    = 10
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The hostname for the MX record (FQDN, e.g., example.com).
- `destination` (String) The mail server hostname.
- `priority` (Number) The priority of the mail server (lower values have higher priority). Must be between 0 and 65535.

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_mx_record.mail example.com/123456
```
