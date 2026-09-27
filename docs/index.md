---
page_title: "Zone.EU Provider"
description: |-
  The Zone.EU provider is used to manage DNS records on Zone.EU hosting platform.
---

# Zone.EU Provider

This provider is used to manage DNS records on [Zone.EU](https://www.zone.ee) hosting platform via their [API v2](https://api.zone.eu/v2).

## Authentication

This provider requires Zone.EU API credentials in order to manage resources.

To manage DNS records, you need:
1. Your ZoneID username
2. An API key generated in your ZoneID account management

There are several ways to provide the required credentials:

* **Set the `username` and `api_key` arguments in the provider configuration**. You can set these arguments in the provider configuration block. Use input variables for the credentials.
* **Set the `ZONE_EU_USERNAME` and `ZONE_EU_API_KEY` environment variables**. The provider can read these environment variables for authentication.

## Example Usage

```terraform
provider "zone" {
  username = var.zone_username
  api_key  = var.zone_api_key
}
```

## Rate Limits

The Zone.EU API has a rate limit of 60 requests per minute per IP address.

## Schema

### Optional

- `username` (String) The ZoneID username used to authenticate with Zone.EU API.
- `api_key` (String) The API key used to authenticate with Zone.EU API.
