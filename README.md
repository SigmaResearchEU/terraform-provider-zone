# Terraform Provider for Zone.EU

Manages DNS records on [Zone.EU](https://www.zone.eu/) through the
[ZoneID API v2](https://api.zone.eu/v2).

```hcl
terraform {
  required_providers {
    zone = {
      source = "sigmaresearcheu/zone"
    }
  }
}

provider "zone" {
  # Or set ZONE_EU_USERNAME / ZONE_EU_API_KEY.
  username = var.zone_username
  api_key  = var.zone_api_key
}

resource "zone_dns_a_record" "www" {
  zone        = "example.com"
  name        = "www.example.com"
  destination = "192.0.2.1"
}
```

Resource and data source reference: [`docs/`](docs/), also rendered on the
Terraform Registry.

## Scope

- **Resources:** all eleven DNS record types — A, AAAA, CAA, CNAME, MX, NS,
  SRV, SSHFP, TLSA, TXT, URL.
- **Data sources:** `zone_dns_zone`, `zone_domain`.

Domain settings, nameserver delegation and anything under `/order/*` are
deliberately out of scope: those endpoints can transfer, renew or buy things.

## API behaviour worth knowing

These were verified against the live API and are why the provider behaves the
way it does.

- **Names are FQDNs.** Write `name = "www.example.com"`, and the zone apex as
  `name = "example.com"`. The API returns FQDNs and rejects short names with
  `422 {"name":"invalid_host"}`.
- **There is no TTL.** Zone.EU fixes TTL server-side; the API has no field
  for it, so neither does this provider.
- **TXT values pass through verbatim.** Zone.EU stores some TXT values with
  literal surrounding quotes and some without. The provider does not normalize
  them; write the value exactly as the API holds it, or every plan will diff.
- **Existing records must be imported.** Creating a record that conflicts with
  one already in the zone fails with `zone_conflict`. Import it instead:
  `terraform import zone_dns_txt_record.spf example.com/<record_id>`.
  Record IDs are shown by `GET /dns/{zone}/{type}`.
- **Rate limit** is 60 requests per minute per IP. The client tracks the
  `X-Ratelimit-*` headers and backs off on 429.

## Development

```sh
make build      # go build
make test       # unit tests (no network)
make docs       # regenerate docs/ from schemas and examples/
make testacc    # acceptance tests — creates and deletes real records
```

Acceptance tests need `ZONE_EU_USERNAME`, `ZONE_EU_API_KEY` and
`ZONE_EU_TEST_DOMAIN`. They create randomly-named `tf-acc-*` records in that
zone and delete them afterwards.

To use a local build, add a `dev_overrides` block for
`sigmaresearcheu/zone` to `~/.terraformrc` pointing at your `$GOPATH/bin`,
then `go install`.

`api/openapi.json` is the OpenAPI spec extracted from the Zone.EU API docs.
It declares record `id` as an integer; the API actually returns strings, and
single-object GETs return one-element arrays. Do not generate a client from it
without accounting for both.

## Releasing

Push a `vX.Y.Z` tag. The release workflow builds with GoReleaser, signs the
checksums with the key in the `GPG_PRIVATE_KEY` / `PASSPHRASE` secrets, and
publishes a GitHub release, which the Terraform Registry picks up.

## License and attribution

MPL-2.0. This provider is a fork of
[kepsic/terraform-provider-zone_eu](https://github.com/kepsic/terraform-provider-zone_eu)
by Andres Kepler; see [NOTICE](NOTICE).
