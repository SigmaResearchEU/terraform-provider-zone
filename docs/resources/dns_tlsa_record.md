---
page_title: "zone_dns_tlsa_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS TLSA record on Zone.EU.
---

# zone_dns_tlsa_record (Resource)

Manages a DNS TLSA record on Zone.EU. TLSA records are used for DANE (DNS-based Authentication of Named Entities) to associate TLS certificates with domain names.

## Example Usage

```terraform
resource "zone_dns_tlsa_record" "https" {
  zone              = "example.com"
  name              = "_443._tcp.example.com"
  destination       = "abc123def456789..."
  certificate_usage = 3
  selector          = 1
  matching_type     = 1
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The service name (e.g., _443._tcp.example.com for HTTPS).
- `destination` (String) The certificate association data (hash or full certificate).
- `certificate_usage` (Number) TLSA certificate usage field. Must be between 0 and 3:
  - 0: CA constraint
  - 1: Service certificate constraint
  - 2: Trust anchor assertion
  - 3: Domain-issued certificate
- `selector` (Number) TLSA selector field. Must be 0 or 1:
  - 0: Full certificate
  - 1: SubjectPublicKeyInfo
- `matching_type` (Number) TLSA matching type field. Must be between 0 and 2:
  - 0: Exact match
  - 1: SHA-256 hash
  - 2: SHA-512 hash

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_tlsa_record.https example.com/123456
```
