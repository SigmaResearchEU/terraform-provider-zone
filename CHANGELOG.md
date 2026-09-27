# Changelog

All notable changes to the Zone.EU Terraform Provider will be documented in this file.

## [Unreleased]

### Added
- Input validation for all DNS record types:
  - A records: IPv4 address validation
  - AAAA records: IPv6 address validation
  - MX records: Priority range validation (0-65535)
  - CAA records: Flag range (0-255) and tag enum validation (`issue`, `issuewild`, `iodef`)
  - SRV records: Priority, weight, and port range validation (0-65535)
  - TLSA records: Certificate usage, selector, and matching type validation
  - SSHFP records: Algorithm and fingerprint type validation
  - URL records: URL format validation and redirect type validation (301, 302)
- Context support in HTTP client for proper cancellation
- ID attribute for `zone_domain` data source
- `FindAllXXXRecordsByName` functions in client for all DNS record types to handle duplicate records

### Fixed
- Delete operations are now idempotent - 404 errors are ignored for already-deleted resources
- Removed incorrect `UseStateForUnknown()` plan modifiers from required mutable fields
- Domain resource now properly handles 404 errors in Read method
- Domain data source now properly sets ID attribute
- **Duplicate DNS records handling**: When `force_recreate = true` and a `zone_conflict` error occurs during update, the provider now:
  - Finds ALL records with the same name (not just the first one)
  - Deletes all duplicate records
  - Creates a fresh record with the desired configuration
  - This fixes the issue where duplicate CNAME, A, AAAA, TXT, MX, NS, SRV, CAA, SSHFP, TLSA, and URL records would cause update failures

### Changed
- HTTP client now uses `context.Context` for request cancellation support
- Update operations for all DNS record types now handle `zone_conflict` errors gracefully when `force_recreate` is enabled

## [1.0.0] - Initial Release

### Added
- Initial release of Zone.EU Terraform Provider
- DNS A record management (`zone_dns_a_record`)
- DNS AAAA record management (`zone_dns_aaaa_record`)
- DNS CNAME record management (`zone_dns_cname_record`)
- DNS MX record management (`zone_dns_mx_record`)
- DNS TXT record management (`zone_dns_txt_record`)
- DNS NS record management (`zone_dns_ns_record`)
- DNS SRV record management (`zone_dns_srv_record`)
- DNS CAA record management (`zone_dns_caa_record`)
- DNS TLSA record management (`zone_dns_tlsa_record`)
- DNS SSHFP record management (`zone_dns_sshfp_record`)
- DNS URL redirect record management (`zone_dns_url_record`)
- DNS Zone data source (`zone_dns_zone`)
- Support for import of existing records
- Authentication via environment variables (`ZONE_EU_USERNAME`, `ZONE_EU_API_KEY`)

### Notes
- Built with Terraform Plugin Framework v1.5.0
- Requires Go 1.21+
- API documentation: https://api.zone.eu/v2
