# Changelog

## Unreleased

First release of the `sigmaresearcheu/zone` provider, forked from
[kepsic/terraform-provider-zone_eu](https://github.com/kepsic/terraform-provider-zone_eu)
at `fbe9057`. Changes from upstream:

### Removed
- `force_recreate` on all DNS record resources. On update it deleted *every*
  record sharing the name, which destroys unrelated records wherever a name
  legitimately holds several (apex TXT, round-robin A, multiple MX).
- Adopting an existing record into state on `zone_conflict`. It matched by
  name and took the first hit, so it could adopt the wrong record. A conflict
  is now an error; use `terraform import`.
- FQDN/short-name normalization in record lookups. Its premise (that the API
  returns short names) is false: the API returns FQDNs and rejects short
  names.
- `zone_domain` and `zone_domain_nameserver` resources. Out of scope: they
  change autorenew, DNSSEC and delegation.

### Changed
- Provider type name is `zone` (resources are `zone_dns_*`), served on plugin
  protocol 6 at `registry.terraform.io/sigmaresearcheu/zone`.
- The API client is one set of generic record methods instead of eleven
  copies, with typed errors: 404 and empty-array responses are "not found",
  and 422 bodies are kept verbatim.
- Rate-limit waits honour context cancellation.

### Fixed
- `zone_dns_zone` failed on every read: the API returns the zone wrapped in an
  array. It now also exposes `active` and `ipv6`, which the example already
  referenced.
- A record that disappears (empty-array GET) is now removed from state rather
  than failing the plan.
- Acceptance tests used a nonexistent `domain` attribute and short names; they
  now use `zone` and random FQDNs, and only run on manual dispatch.
- Client tests now exercise the client against an `httptest` server.
- The release produces the `_manifest.json` the Terraform Registry requires.
