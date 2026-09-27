---
page_title: "zone_dns_srv_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS SRV record on Zone.EU.
---

# zone_dns_srv_record (Resource)

Manages a DNS SRV record on Zone.EU. SRV records specify the location of servers for specific services.

## Example Usage

```terraform
resource "zone_dns_srv_record" "sip" {
  zone        = "example.com"
  name        = "_sip._tcp.example.com"
  destination = "sipserver.example.com"
  priority    = 10
  weight      = 5
  port        = 5060
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The service name (e.g., _sip._tcp.example.com).
- `destination` (String) The target server hostname.
- `priority` (Number) The priority of the target host (lower values have higher priority). Must be between 0 and 65535.
- `weight` (Number) A relative weight for records with the same priority. Must be between 0 and 65535.
- `port` (Number) The TCP or UDP port on which the service is found. Must be between 0 and 65535.

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_srv_record.sip example.com/123456
```
