package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccName returns a random FQDN under the test zone. The API rejects short
// names with 422 invalid_host, so acceptance configs always use FQDNs.
func testAccName(t *testing.T, zone string) string {
	t.Helper()
	return fmt.Sprintf("tf-acc-%s.%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum), zone)
}

func testAccZone(t *testing.T) string {
	t.Helper()
	zone := os.Getenv("ZONE_EU_TEST_DOMAIN")
	if zone == "" {
		t.Skip("ZONE_EU_TEST_DOMAIN must be set for acceptance tests")
	}
	return zone
}

func TestAccDNSARecordResource(t *testing.T) {
	zone := testAccZone(t)
	name := testAccName(t, zone)
	resourceName := "zone_dns_a_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordConfig("zone_dns_a_record", zone, name, "192.0.2.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone", zone),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "destination", "192.0.2.1"),
					resource.TestCheckResourceAttrSet(resourceName, "record_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDNSRecordConfig("zone_dns_a_record", zone, name, "192.0.2.2"),
				Check:  resource.TestCheckResourceAttr(resourceName, "destination", "192.0.2.2"),
			},
		},
	})
}

func TestAccDNSTXTRecordResource(t *testing.T) {
	zone := testAccZone(t)
	name := testAccName(t, zone)
	resourceName := "zone_dns_txt_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordConfig("zone_dns_txt_record", zone, name, "v=spf1 -all"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "destination", "v=spf1 -all"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDNSRecordConfig(resourceType, zone, name, destination string) string {
	return fmt.Sprintf(`
resource %[1]q "test" {
  zone        = %[2]q
  name        = %[3]q
  destination = %[4]q
}
`, resourceType, zone, name, destination)
}
