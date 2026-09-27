---
page_title: "zone_dns_sshfp_record Resource - terraform-provider-zone"
subcategory: ""
description: |-
  Manages a DNS SSHFP record on Zone.EU.
---

# zone_dns_sshfp_record (Resource)

Manages a DNS SSHFP record on Zone.EU. SSHFP records store SSH public key fingerprints for host key verification.

## Example Usage

```terraform
resource "zone_dns_sshfp_record" "server" {
  zone        = "example.com"
  name        = "server.example.com"
  destination = "abc123def456789..."
  algorithm   = 4
  type        = 2
}
```

## Schema

### Required

- `zone` (String) The DNS zone name (domain name, e.g., example.com).
- `name` (String) The hostname for the SSHFP record (FQDN, e.g., server.example.com).
- `destination` (String) The fingerprint in hexadecimal.
- `algorithm` (Number) The SSH key algorithm. Must be between 1 and 4:
  - 1: RSA
  - 2: DSA
  - 3: ECDSA
  - 4: Ed25519
- `fingerprint_type` (Number) The fingerprint type. Must be 1 or 2:
  - 1: SHA-1
  - 2: SHA-256

### Optional

- `force_recreate` (Boolean) If true, updates an existing record with the same name instead of creating a new one. Default: `false`.

### Read-Only

- `id` (String) The ID of this resource in format `zone/record_id`.
- `record_id` (String) The ID of the record in Zone.EU.

## Import

Import is supported using the following syntax:

```shell
terraform import zone_dns_sshfp_record.server example.com/123456
```
