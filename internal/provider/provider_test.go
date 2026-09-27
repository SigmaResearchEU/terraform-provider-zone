package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"zone": providerserver.NewProtocol6WithError(New("test")()),
}

// TestProviderSchema validates the provider, resource and data source schemas
// the way Terraform does when it loads the provider.
func TestProviderSchema(t *testing.T) {
	server, err := testAccProtoV6ProviderFactories["zone"]()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range resp.Diagnostics {
		t.Errorf("%s: %s", d.Summary, d.Detail)
	}

	wantResources := []string{
		"zone_dns_a_record", "zone_dns_aaaa_record", "zone_dns_caa_record",
		"zone_dns_cname_record", "zone_dns_mx_record", "zone_dns_ns_record",
		"zone_dns_srv_record", "zone_dns_sshfp_record", "zone_dns_tlsa_record",
		"zone_dns_txt_record", "zone_dns_url_record",
	}
	if len(resp.ResourceSchemas) != len(wantResources) {
		t.Errorf("expected %d resources, got %d", len(wantResources), len(resp.ResourceSchemas))
	}
	for _, name := range wantResources {
		s, ok := resp.ResourceSchemas[name]
		if !ok {
			t.Errorf("missing resource %s", name)
			continue
		}
		for _, attr := range s.Block.Attributes {
			if attr.Name == "force_recreate" || attr.Name == "ttl" {
				t.Errorf("%s must not expose %q", name, attr.Name)
			}
		}
	}
	for _, name := range []string{"zone_dns_zone", "zone_domain"} {
		if _, ok := resp.DataSourceSchemas[name]; !ok {
			t.Errorf("missing data source %s", name)
		}
	}
}
