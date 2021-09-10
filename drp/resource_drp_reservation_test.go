package drp

import (
	"fmt"
	"net"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpReservation_basic = `
	resource "drp_reservation" "foo" {
		Addr = "1.1.1.1"
		Token = "aa:bb:cc:dd:ee:ff"
		Strategy = "MAC"
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpReservation_basic(t *testing.T) {
	reservation := models.Reservation{
		Addr:     net.ParseIP("1.1.1.1"),
		Token:    "aa:bb:cc:dd:ee:ff",
		Strategy: "MAC",
		Meta:     map[string]string{"field1": "value1", "field2": "value2"},
	}

	reservation.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckReservationDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpReservation_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckReservationExists(t, "drp_reservation.foo", &reservation),
				),
			},
		},
	})
}

var testAccDrpReservation_change_1 = `
	resource "drp_reservation" "foo" {
		Addr = "1.1.1.1"
		Token = "aa:bb:cc:dd:ee:ff"
		Strategy = "MAC"
		NextServer = "1.2.3.4"
		Options = [
			{ Code = 30, Value = "fred" },
			{ Code = 3, Value = "1.1.1.1" }
		]
	}`

var testAccDrpReservation_change_2 = `
	resource "drp_reservation" "foo" {
		Addr = "1.1.1.1"
		Token = "aa:bb:cc:dd:ee:ff"
		Strategy = "MAC"
		NextServer = "1.2.3.5"
		Options = [
			{ Code = 33, Value = "fred" },
			{ Code = 4, Value = "1.2.1.1" }
		]
	}`

func TestAccDrpReservation_change(t *testing.T) {
	reservation1 := models.Reservation{
		Addr:       net.ParseIP("1.1.1.1"),
		Token:      "aa:bb:cc:dd:ee:ff",
		Strategy:   "MAC",
		NextServer: net.ParseIP("1.2.3.4"),
		Options: []models.DhcpOption{
			models.DhcpOption{Code: 30, Value: "fred"},
			models.DhcpOption{Code: 3, Value: "1.1.1.1"},
		},
	}
	reservation1.Fill()
	reservation2 := models.Reservation{
		Addr:       net.ParseIP("1.1.1.1"),
		Token:      "aa:bb:cc:dd:ee:ff",
		Strategy:   "MAC",
		NextServer: net.ParseIP("1.2.3.5"),
		Options: []models.DhcpOption{
			models.DhcpOption{Code: 33, Value: "fred"},
			models.DhcpOption{Code: 4, Value: "1.2.1.1"},
		},
	}
	reservation2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckReservationDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpReservation_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckReservationExists(t, "drp_reservation.foo", &reservation1),
				),
			},
			resource.TestStep{
				Config: testAccDrpReservation_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckReservationExists(t, "drp_reservation.foo", &reservation2),
				),
			},
		},
	})
}

func testAccDrpCheckReservationDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_reservation" {
			continue
		}

		if _, err := config.session.GetModel("reservations", rs.Primary.ID); err == nil {
			return fmt.Errorf("Reservation still exists")
		}
	}

	return nil
}

func testAccDrpCheckReservationExists(t *testing.T, n string, reservation *models.Reservation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("reservations", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.Reservation)
		found.ClearValidation()

		if found.Key() != rs.Primary.ID {
			return fmt.Errorf("Reservation not found: %s %s", found.Key(), rs.Primary.ID)
		}

		if err := diffObjects(reservation, found, "Reservation"); err != nil {
			return err
		}
		return nil
	}
}
