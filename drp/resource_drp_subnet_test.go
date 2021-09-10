package drp

import (
	"fmt"
	"net"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpSubnet_basic = `
	resource "drp_subnet" "foo" {
		Name = "foo"
		Subnet = "1.1.1.0/24"
		Strategy = "MAC"
		ActiveStart = "1.1.1.4"
		ActiveEnd = "1.1.1.9"
		ActiveLeaseTime = 120
		ReservedLeaseTime = 14400
		Options = []
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpSubnet_basic(t *testing.T) {
	subnet := models.Subnet{Name: "foo",
		ActiveLeaseTime:   120,
		ReservedLeaseTime: 14400,
		Subnet:            "1.1.1.0/24",
		ActiveStart:       net.ParseIP("1.1.1.4"),
		ActiveEnd:         net.ParseIP("1.1.1.9"),
		Strategy:          "MAC",
		Meta:              map[string]string{"field1": "value1", "field2": "value2"},
		Pickers:           []string{"hint", "nextFree", "mostExpired"},
		Options: []models.DhcpOption{
			models.DhcpOption{Code: 1, Value: "255.255.255.0"},
			models.DhcpOption{Code: 28, Value: "1.1.1.255"},
		},
	}

	subnet.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckSubnetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpSubnet_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckSubnetExists(t, "drp_subnet.foo", &subnet),
				),
			},
		},
	})
}

var testAccDrpSubnet_change_1 = `
	resource "drp_subnet" "foo" {
		Name = "foo"
		Enabled = true
		Proxy = true
		Subnet = "1.1.1.0/24"
		NextServer = "1.1.1.1"
		ActiveStart = "1.1.1.4"
		ActiveEnd = "1.1.1.9"
		ActiveLeaseTime = 120
		ReservedLeaseTime = 14400
		OnlyReservations = true
		Strategy = "MAC"
		Options = [
			{ Code = 30, Value = "fred" },
			{ Code = 3, Value = "1.1.1.1" }
		]
		Pickers = [ "none" ]
	}`

var testAccDrpSubnet_change_2 = `
	resource "drp_subnet" "foo" {
		Name = "foo"
		Enabled = false
		Proxy = false
		Subnet = "1.1.1.0/24"
		NextServer = "1.1.1.2"
		ActiveStart = "1.1.1.5"
		ActiveEnd = "1.1.1.10"
		ActiveLeaseTime = 121
		ReservedLeaseTime = 14401
		OnlyReservations = false
		Strategy = "MAC"
		Options = [
			{ Code = 33, Value = "fred" },
			{ Code = 4, Value = "1.2.1.1" }
		]
		Pickers = [ "hint", "none" ]
	}`

func TestAccDrpSubnet_change(t *testing.T) {
	subnet1 := models.Subnet{Name: "foo",
		ActiveLeaseTime:   120,
		Enabled:           true,
		Proxy:             true,
		OnlyReservations:  true,
		ReservedLeaseTime: 14400,
		Subnet:            "1.1.1.0/24",
		NextServer:        net.ParseIP("1.1.1.1"),
		ActiveStart:       net.ParseIP("1.1.1.4"),
		ActiveEnd:         net.ParseIP("1.1.1.9"),
		Strategy:          "MAC",
		Pickers:           []string{"none"},
		Options: []models.DhcpOption{
			models.DhcpOption{Code: 30, Value: "fred"},
			models.DhcpOption{Code: 3, Value: "1.1.1.1"},
			models.DhcpOption{Code: 1, Value: "255.255.255.0"},
			models.DhcpOption{Code: 28, Value: "1.1.1.255"},
		},
	}
	subnet1.Fill()
	subnet2 := models.Subnet{Name: "foo",
		ActiveLeaseTime:   121,
		ReservedLeaseTime: 14401,
		Subnet:            "1.1.1.0/24",
		NextServer:        net.ParseIP("1.1.1.2"),
		ActiveStart:       net.ParseIP("1.1.1.5"),
		ActiveEnd:         net.ParseIP("1.1.1.10"),
		Strategy:          "MAC",
		Pickers:           []string{"hint", "none"},
		Options: []models.DhcpOption{
			models.DhcpOption{Code: 33, Value: "fred"},
			models.DhcpOption{Code: 4, Value: "1.2.1.1"},
			models.DhcpOption{Code: 1, Value: "255.255.255.0"},
			models.DhcpOption{Code: 28, Value: "1.1.1.255"},
		},
	}
	subnet2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckSubnetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpSubnet_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckSubnetExists(t, "drp_subnet.foo", &subnet1),
				),
			},
			resource.TestStep{
				Config: testAccDrpSubnet_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckSubnetExists(t, "drp_subnet.foo", &subnet2),
				),
			},
		},
	})
}

func testAccDrpCheckSubnetDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_subnet" {
			continue
		}

		if _, err := config.session.GetModel("subnets", rs.Primary.ID); err == nil {
			return fmt.Errorf("Subnet still exists")
		}
	}

	return nil
}

func testAccDrpCheckSubnetExists(t *testing.T, n string, subnet *models.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("subnets", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.Subnet)
		found.ClearValidation()

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("Subnet not found")
		}

		if err := diffObjects(subnet, found, "Subnet"); err != nil {
			return err
		}
		return nil
	}
}
