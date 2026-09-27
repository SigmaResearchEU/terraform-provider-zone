# Fix Zone.eu Terraform Provider: Handle zone_conflict Errors

## Problem

When Terraform tries to create a DNS record that already exists in Zone.eu, the API returns a 422 error with `{"name":"zone_conflict"}`. Currently the provider just errors out, leaving Terraform state out of sync with reality.

```
Error: Client Error

Unable to create CNAME record, got error: API error (status 422):
{"name":"zone_conflict"} (X-Status-Message: Puudulik sisend)
```

## Root Cause Analysis

The current fix in commit `3bdb609` added zone_conflict handling but it's NOT WORKING because of a **name format mismatch**:

1. Terraform passes FQDN: `name = "caddy.vermare.ee"`
2. Zone.eu API returns short name: `name = "caddy"` (without zone suffix)
3. The comparison `r.Name == name` fails because `"caddy" != "caddy.vermare.ee"`

## Required Fix

### Step 1: Fix FindXXXRecordByName functions in client.go

The `FindXXXRecordByName` functions need to handle both FQDN and short name formats:

```go
func (c *Client) FindCNAMERecordByName(zone, name string) (*DNSRecord, error) {
    records, err := c.ListCNAMERecords(zone)
    if err != nil {
        return nil, err
    }
    
    // Normalize the search name - strip zone suffix if present
    searchName := name
    zoneSuffix := "." + zone
    if strings.HasSuffix(name, zoneSuffix) {
        searchName = strings.TrimSuffix(name, zoneSuffix)
    }
    
    for _, r := range records {
        // Compare both the record name and normalized search name
        recordName := r.Name
        if strings.HasSuffix(r.Name, zoneSuffix) {
            recordName = strings.TrimSuffix(r.Name, zoneSuffix)
        }
        
        if recordName == searchName || r.Name == name {
            return &r, nil
        }
    }
    return nil, nil // Not found
}
```

Apply this same pattern to ALL FindXXXRecordByName functions:
- `FindARecordByName`
- `FindAAAARecordByName`
- `FindCNAMERecordByName`
- `FindTXTRecordByName`
- `FindMXRecordByName`
- `FindNSRecordByName`
- `FindSRVRecordByName`
- `FindCAARecordByName`
- `FindSSHFPRecordByName`
- `FindTLSARecordByName`
- `FindURLRecordByName`

### Step 2: The Create function zone_conflict handling (already done)

The Create functions already have zone_conflict handling from commit `3bdb609`. Once Step 1 is fixed, they will work correctly.

## Files to Modify

**client.go** - Fix all FindXXXRecordByName functions to handle name format mismatch

All DNS resource files already have zone_conflict handling:

- `internal/provider/resource_dns_a_record.go`
- `internal/provider/resource_dns_aaaa_record.go`
- `internal/provider/resource_dns_cname_record.go`
- `internal/provider/resource_dns_txt_record.go`
- `internal/provider/resource_dns_mx_record.go`
- `internal/provider/resource_dns_ns_record.go`
- `internal/provider/resource_dns_srv_record.go`
- `internal/provider/resource_dns_caa_record.go`
- `internal/provider/resource_dns_sshfp_record.go`
- `internal/provider/resource_dns_tlsa_record.go`
- `internal/provider/resource_dns_url_record.go`

## Testing

After making changes:

```bash
# Build the provider
cd /Users/andrke/dev/terraform-provider-zone_eu
go build -o terraform-provider-zoneeu

# The dev_overrides in ~/.terraformrc will use this build
# Test with terraform
cd /Users/andrke/dev/paplirahu/infrastructure/environments/production
terraform apply
```

## Expected Behavior After Fix

When a record already exists:

```
module.tenant["ee"].zoneeu_dns_cname_record.app["app.vermare.ee"]: Creating...
2026-01-28T17:00:00.000Z [INFO] Record already exists (zone_conflict), adopting into state
module.tenant["ee"].zoneeu_dns_cname_record.app["app.vermare.ee"]: Creation complete after 1s

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

## Debug Info

Current state shows name format:
```hcl
# Working record in state:
resource "zoneeu_dns_cname_record" "www" {
    name           = "www.vermare.lv"   # FQDN format
    zone           = "vermare.lv"
}
```

But Zone.eu API might return:
```json
{"name": "www", "destination": "..."}  // Short format without zone
```

This is why the comparison fails and the fix doesn't work.
