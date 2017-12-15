package drp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/digitalrebar/provision/models"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var testAccDrpMachine_basic = `
	resource "drp_machine" "foo" {
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpMachine_basic(t *testing.T) {
	machine := models.Machine{Name: "mach1",
		Meta: map[string]string{"field1": "value1", "field2": "value2"},
	}
	machine.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckMachineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpMachine_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckMachineExists(t, "drp_machine.foo", &machine),
				),
			},
		},
	})
}

var testAccDrpMachine_change_1 = `
	resource "drp_machine" "foo" {
		Name = "foo"
		Description = "I am a machine"
	}`

var testAccDrpMachine_change_2 = `
	resource "drp_machine" "foo" {
		Name = "foo"
		Description = "I am a machine again"
	}`

func TestAccDrpMachine_change(t *testing.T) {
	machine1 := models.Machine{Name: "foo", Description: "I am a machine"}
	machine1.Fill()
	machine2 := models.Machine{Name: "foo", Description: "I am a machine again"}
	machine2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckMachineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpMachine_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckMachineExists(t, "drp_machine.foo", &machine1),
				),
			},
			resource.TestStep{
				Config: testAccDrpMachine_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckMachineExists(t, "drp_machine.foo", &machine2),
				),
			},
		},
	})
}

var testAccDrpMachine_withParams = `
	resource "drp_machine" "foo" {
		Name = "foo"
		Description = "I am a machine again"
		Params = {
			"test/string" = "fred"
			"test/int" = 3
			"test/bool" = true
			"test/list" = [ "one", "two" ]
		}
	}`

func TestAccDrpMachine_withParams(t *testing.T) {
	machine := models.Machine{Name: "foo", Description: "I am a machine",
		Params: map[string]interface{}{
			"test/string": "fred",
			"test/int":    3,
			"test/bool":   true,
			"test/list":   []string{"one", "two"},
		},
	}
	machine.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckMachineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpMachine_withParams,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckMachineExists(t, "drp_machine.foo", &machine),
				),
			},
		},
	})
}

func testAccDrpCheckMachineDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_machine" {
			continue
		}

		if _, err := config.session.GetModel("machines", rs.Primary.ID); err == nil {
			return fmt.Errorf("Machine still exists")
		}
	}

	return nil
}

func testAccDrpCheckMachineExists(t *testing.T, n string, machine *models.Machine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("machines", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.Machine)
		found.ClearValidation()

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("Machine not found")
		}

		if !reflect.DeepEqual(machine, found) {
			b1, _ := json.MarshalIndent(machine, "", "  ")
			b2, _ := json.MarshalIndent(found, "", "  ")
			return fmt.Errorf("Machine doesn't match: e:%s\na:%s", string(b1), string(b2))
		}
		return nil
	}
}
