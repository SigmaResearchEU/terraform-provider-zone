# Terraform Provider for Zone.EU

This is a Terraform provider for managing DNS records and domains on [Zone.EU](https://www.zone.ee) hosting platform via their [API v2](https://api.zone.eu/v2).

## Features

### Implemented

#### DNS Management
- **DNS A Record** - IPv4 address records
- **DNS AAAA Record** - IPv6 address records
- **DNS CNAME Record** - Canonical name (alias) records
- **DNS MX Record** - Mail exchange records
- **DNS TXT Record** - Text records (SPF, DKIM, etc.)
- **DNS NS Record** - Nameserver records
- **DNS SRV Record** - Service locator records
- **DNS CAA Record** - Certificate Authority Authorization records
- **DNS TLSA Record** - DANE/TLS authentication records
- **DNS SSHFP Record** - SSH fingerprint records
- **DNS URL Record** - URL redirect records (Zone.EU specific)

#### Domain Management
- **Domain** - Manage domain settings (autorenew, DNSSEC, renewal notifications, custom nameservers)
- **Domain Nameserver** - Manage custom nameservers for domains

### Data Sources

- **DNS Zone** - Read DNS zone information
- **Domain** - Read domain information

### Not Yet Implemented

The Zone.EU API supports many other services that are not yet implemented in this provider:

- **Domain Registration/Transfer** - Domain registration is not available via API
- **Domain Contacts** - Contact management for domains
- **Webhosting (vserver)** - Virtual server management
- **E-mail** - Email account management
- **MySQL** - Database management
- **SSL Certificates** - SSL/TLS certificate management
- **Crontab** - Scheduled task management
- **Redis** - Redis database management
- **PM2** - Node.js process management
- **SSH** - SSH key and whitelist management
- **Port Forward** - Port forwarding configuration
- **Cloudserver VPS** - Cloud VPS management

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.12
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- Zone.EU account with API access enabled

## Installation

### Building from source

```bash
git clone https://github.com/kepsic/terraform-provider-zone_eu.git
cd terraform-provider-zone_eu
go build -o terraform-provider-zoneeu
```

### Installing locally

Copy the built binary to your Terraform plugins directory:

```bash
# Linux/macOS
mkdir -p ~/.terraform.d/plugins
cp terraform-provider-zoneeu ~/.terraform.d/plugins/

# Or for Terraform 0.13+
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/kepsic/zoneeu/1.0.0/darwin_amd64/
cp terraform-provider-zoneeu ~/.terraform.d/plugins/registry.terraform.io/kepsic/zoneeu/1.0.0/darwin_amd64/
```

## Configuration

### Provider Configuration

```hcl
terraform {
  required_providers {
    zoneeu = {
      source = "kepsic/zoneeu"
    }
  }
}

provider "zoneeu" {
  username = "your-zoneid-username"
  api_key  = "your-zoneid-api-key"
}
```

You can also use environment variables:

```bash
export ZONE_EU_USERNAME="your-zoneid-username"
export ZONE_EU_API_KEY="your-zoneid-api-key"
```

### Authentication

The provider uses HTTP Basic Auth. You need:
1. Your ZoneID username
2. An API key generated in your ZoneID account management

## Usage Examples

### A Record

```hcl
resource "zoneeu_dns_a_record" "www" {
  zone        = "example.com"
  name        = "www.example.com"
  destination = "192.168.1.1"
}
```

### AAAA Record (IPv6)

```hcl
resource "zoneeu_dns_aaaa_record" "www" {
  zone        = "example.com"
  name        = "www.example.com"
  destination = "2001:db8::1"
}
```

### CNAME Record

```hcl
resource "zoneeu_dns_cname_record" "blog" {
  zone        = "example.com"
  name        = "blog.example.com"
  destination = "bloghost.example.net"
}
```

### MX Record

```hcl
resource "zoneeu_dns_mx_record" "mail" {
  zone        = "example.com"
  name        = "example.com"
  destination = "mail.example.com"
  priority    = 10
}
```

### TXT Record (SPF)

```hcl
resource "zoneeu_dns_txt_record" "spf" {
  zone        = "example.com"
  name        = "example.com"
  destination = "v=spf1 include:_spf.google.com ~all"
}
```

### NS Record

```hcl
resource "zoneeu_dns_ns_record" "subdomain" {
  zone        = "example.com"
  name        = "subdomain.example.com"
  destination = "ns1.otherdns.com"
}
```

### SRV Record

```hcl
resource "zoneeu_dns_srv_record" "sip" {
  zone        = "example.com"
  name        = "_sip._tcp.example.com"
  destination = "sipserver.example.com"
  priority    = 10
  weight      = 5
  port        = 5060
}
```

### CAA Record

```hcl
resource "zoneeu_dns_caa_record" "letsencrypt" {
  zone        = "example.com"
  name        = "example.com"
  destination = "letsencrypt.org"
  flag        = 0
  tag         = "issue"
}
```

### TLSA Record (DANE)

```hcl
resource "zoneeu_dns_tlsa_record" "https" {
  zone              = "example.com"
  name              = "_443._tcp.example.com"
  destination       = "abc123def456..."
  certificate_usage = 3
  selector          = 1
  matching_type     = 1
}
```

### SSHFP Record

```hcl
resource "zoneeu_dns_sshfp_record" "server" {
  zone        = "example.com"
  name        = "server.example.com"
  destination = "abc123def456..."
  algorithm   = 4  # Ed25519
  type        = 2  # SHA-256
}
```

### URL Redirect Record

```hcl
resource "zoneeu_dns_url_record" "redirect" {
  zone          = "example.com"
  name          = "old.example.com"
  destination   = "https://new.example.com/page"
  redirect_type = 301
}
```

### Data Source: DNS Zone

```hcl
data "zoneeu_dns_zone" "main" {
  name = "example.com"
}

output "zone_active" {
  value = data.zoneeu_dns_zone.main.active
}
```

### Domain Resource

Manage settings for an existing domain:

```hcl
resource "zoneeu_domain" "example" {
  name                   = "example.com"
  autorenew              = true
  dnssec                 = true
  renewal_notifications  = true
  nameservers_custom     = false
}
```

### Data Source: Domain

```hcl
data "zoneeu_domain" "example" {
  name = "example.com"
}

output "domain_expires" {
  value = data.zoneeu_domain.example.expires
}
```

### Custom Nameservers

To use custom nameservers, first enable them on the domain, then add the nameserver records:

```hcl
resource "zoneeu_domain" "example" {
  name               = "example.com"
  nameservers_custom = true
}

resource "zoneeu_domain_nameserver" "ns1" {
  domain   = zoneeu_domain.example.name
  hostname = "ns1.example.com"
  ip       = ["192.168.1.1"]  # Glue record - required when NS is under same domain
}

resource "zoneeu_domain_nameserver" "ns2" {
  domain   = zoneeu_domain.example.name
  hostname = "ns2.example.com"
  ip       = ["192.168.1.2"]
}

# External nameserver (no glue record needed)
resource "zoneeu_domain_nameserver" "external" {
  domain   = zoneeu_domain.example.name
  hostname = "ns1.externaldns.com"
}
```

## Importing Existing Resources

If you have existing DNS records or domains in Zone.EU that you want to manage with Terraform, you need to import them into your Terraform state first. Otherwise, Terraform will try to create new records and fail with a `zone_conflict` error.

### Finding Record IDs

You can find record IDs using the Zone.EU API:

```bash
# List all A records for a zone
curl -u "username:apikey" https://api.zone.eu/v2/dns/example.com/a

# List all CNAME records
curl -u "username:apikey" https://api.zone.eu/v2/dns/example.com/cname

# List all records of any type (replace 'a' with: aaaa, cname, mx, txt, ns, srv, caa, tlsa, sshfp, url)
curl -u "username:apikey" https://api.zone.eu/v2/dns/example.com/{type}
```

The response will include the `id` field for each record.

### Import Commands

#### DNS Records

```bash
# Format: terraform import RESOURCE_ADDRESS ZONE/RECORD_ID

# A Record
terraform import zoneeu_dns_a_record.www example.com/123

# AAAA Record
terraform import zoneeu_dns_aaaa_record.www example.com/456

# CNAME Record
terraform import zoneeu_dns_cname_record.blog example.com/789

# MX Record
terraform import zoneeu_dns_mx_record.mail example.com/101

# TXT Record
terraform import zoneeu_dns_txt_record.spf example.com/102

# NS Record
terraform import zoneeu_dns_ns_record.subdomain example.com/103

# SRV Record
terraform import zoneeu_dns_srv_record.sip example.com/104

# CAA Record
terraform import zoneeu_dns_caa_record.letsencrypt example.com/105

# TLSA Record
terraform import zoneeu_dns_tlsa_record.https example.com/106

# SSHFP Record
terraform import zoneeu_dns_sshfp_record.server example.com/107

# URL Record
terraform import zoneeu_dns_url_record.redirect example.com/108
```

#### Importing with for_each

When using `for_each`, use quotes around the resource address:

```bash
# Example: importing into a for_each resource
terraform import 'zoneeu_dns_a_record.root["example.com"]' example.com/123
terraform import 'module.tenant["lv"].zoneeu_dns_a_record.root["vermare.lv"]' vermare.lv/456
terraform import 'module.tenant["ee"].zoneeu_dns_cname_record.caddy["caddy.vermare.ee"]' vermare.ee/789
```

#### Domain

```bash
# Format: domain_name
terraform import zoneeu_domain.example example.com
```

#### Domain Nameserver

```bash
# Format: domain/hostname
terraform import zoneeu_domain_nameserver.ns1 example.com/ns1.example.com
```

### Common Import Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `zone_conflict` | Record already exists | Import the existing record or use `force_recreate = true` |
| `not found` | Wrong record ID or zone | Verify the ID using the API |
| `invalid import id` | Wrong format | Use `zone/record_id` format |

## Force Recreate Option

All DNS record resources support the `force_recreate` attribute as an alternative to importing existing records. When set to `true`, the provider will automatically delete any existing record with the same name before creating the new one.

### Usage

```hcl
resource "zoneeu_dns_a_record" "www" {
  zone           = "example.com"
  name           = "www.example.com"
  destination    = "192.168.1.1"
  force_recreate = true  # Delete existing record with same name if it exists
}
```

### When to Use

| Scenario | Recommended Approach |
|----------|---------------------|
| Taking over existing infrastructure | Import existing records |
| Clean slate / fresh setup | Use `force_recreate = true` |
| Migrating from another DNS provider | Use `force_recreate = true` |
| CI/CD pipelines with ephemeral state | Use `force_recreate = true` |
| Production with state preservation | Import existing records |

### Important Notes

- **Default is `false`**: By default, the provider will not delete existing records
- **Only affects Create**: The `force_recreate` flag only applies during resource creation
- **Matches by name**: The provider finds existing records by matching the `name` field
- **First match only**: If multiple records exist with the same name, only the first one is deleted
- **Data loss warning**: Use with caution - this will delete existing records!

### Example with Multiple Records

```hcl
locals {
  dns_records = {
    www  = "192.168.1.1"
    api  = "192.168.1.2"
    mail = "192.168.1.3"
  }
}

resource "zoneeu_dns_a_record" "records" {
  for_each       = local.dns_records
  zone           = "example.com"
  name           = "${each.key}.example.com"
  destination    = each.value
  force_recreate = true
}
```

## API Rate Limits

The Zone.EU API has a rate limit of 60 requests per minute per IP address.

The provider automatically handles rate limiting:

- **Tracks rate limit headers**: Reads `X-Ratelimit-Limit` and `X-Ratelimit-Remaining` from API responses
- **Automatic retry**: When rate limited (HTTP 429), the provider waits and retries automatically (up to 3 times)
- **Respects Retry-After**: Uses the `Retry-After` header when provided, defaults to 60 seconds

| HTTP Code | Description |
|-----------|-------------|
| 200 | Successful GET request |
| 201 | Successful POST/PUT request |
| 204 | Successful DELETE request |
| 422 | Validation errors in request |
| 429 | Rate limit exceeded (auto-retry) |

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.21+ is *required*).

To compile the provider, run:

```sh
go build -o terraform-provider-zoneeu
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This provider is distributed under the [Mozilla Public License 2.0](LICENSE).
