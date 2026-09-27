package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSARecordResource(t *testing.T) {
	domain := os.Getenv("ZONE_EU_TEST_DOMAIN")
	if domain == "" {
		t.Skip("ZONE_EU_TEST_DOMAIN must be set for acceptance tests")
	}

	resourceName := "zoneeu_dns_a_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDNSARecordResourceConfig(domain, "test-acc", "192.168.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttr(resourceName, "name", "test-acc"),
					resource.TestCheckResourceAttr(resourceName, "destination", "192.168.1.1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_recreate"},
			},
			// Update testing
			{
				Config: testAccDNSARecordResourceConfig(domain, "test-acc", "192.168.1.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "destination", "192.168.1.2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDNSARecordResourceConfig(domain, name, destination string) string {
	return fmt.Sprintf(`
resource "zoneeu_dns_a_record" "test" {
  domain      = %[1]q
  name        = %[2]q
  destination = %[3]q
}
`, domain, name, destination)
}

func TestAccDNSTXTRecordResource(t *testing.T) {
	domain := os.Getenv("ZONE_EU_TEST_DOMAIN")
	if domain == "" {
		t.Skip("ZONE_EU_TEST_DOMAIN must be set for acceptance tests")
	}

	resourceName := "zoneeu_dns_txt_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDNSTXTRecordResourceConfig(domain, "test-acc-txt", "v=spf1 include:test.com ~all"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttr(resourceName, "name", "test-acc-txt"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDNSTXTRecordResourceConfig(domain, name, destination string) string {
	return fmt.Sprintf(`
resource "zoneeu_dns_txt_record" "test" {
  domain      = %[1]q
  name        = %[2]q
  destination = %[3]q
}
`, domain, name, destination)
}

func TestAccDNSARecordResource_ForceRecreate(t *testing.T) {
	domain := os.Getenv("ZONE_EU_TEST_DOMAIN")
	if domain == "" {
		t.Skip("ZONE_EU_TEST_DOMAIN must be set for acceptance tests")
	}

	resourceName := "zoneeu_dns_a_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSARecordResourceConfigForceRecreate(domain, "test-force", "192.168.1.100"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "force_recreate", "true"),
					resource.TestCheckResourceAttr(resourceName, "destination", "192.168.1.100"),
				),
			},
		},
	})
}

func testAccDNSARecordResourceConfigForceRecreate(domain, name, destination string) string {
	return fmt.Sprintf(`
resource "zoneeu_dns_a_record" "test" {
  domain         = %[1]q
  name           = %[2]q
  destination    = %[3]q
  force_recreate = true
}
`, domain, name, destination)
}
