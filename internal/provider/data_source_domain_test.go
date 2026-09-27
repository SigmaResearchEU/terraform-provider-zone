package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccPreCheck validates required environment variables for running acceptance tests.
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ZONE_EU_USERNAME"); v == "" {
		t.Fatal("ZONE_EU_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("ZONE_EU_API_KEY"); v == "" {
		t.Fatal("ZONE_EU_API_KEY must be set for acceptance tests")
	}
}

func TestAccDomainDataSource(t *testing.T) {
	domain := os.Getenv("ZONE_EU_TEST_DOMAIN")
	if domain == "" {
		t.Skip("ZONE_EU_TEST_DOMAIN must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDataSourceConfig(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.zoneeu_domain.test", "name", domain),
					resource.TestCheckResourceAttrSet("data.zoneeu_domain.test", "expires"),
				),
			},
		},
	})
}

func testAccDomainDataSourceConfig(domain string) string {
	return `
data "zoneeu_domain" "test" {
  name = "` + domain + `"
}
`
}
