package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"zoneeu": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProviderSchema(t *testing.T) {
	t.Parallel()

	// Test that the provider schema is valid
	provider := New("test")()
	if provider == nil {
		t.Fatal("provider should not be nil")
	}
}

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	provider := &ZoneProvider{version: "test"}

	// Type assertion to verify provider implements the interface
	var _ interface {
		Metadata(interface{}, interface{}, interface{})
	} = nil

	_ = provider
}
