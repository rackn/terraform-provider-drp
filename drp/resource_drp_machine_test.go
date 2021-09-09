package drp

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pborman/uuid"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpMachine_basic = `
	resource "drp_machine" "foo" {
		Name = "mach1"
		Stage = "local"
		completion_stage = "local"
		decommission_stage = "none"
		add_profiles = [ "p-test" ]
		Meta = {
                        "feature-flags" = "change-stage-v2"
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpMachine_basic(t *testing.T) {
	machine := models.Machine{
		Name:     "mach1",
		Uuid:     uuid.Parse("3945838b-be8c-4b35-8b1c-b538ddc71f7e"),
		Secret:   "12",
		Runnable: true,
		BootEnv:  "local",
		Stage:    "local",
		Profiles: []string{"p-test"},
		Params: map[string]interface{}{
			"terraform/allocated": true,
			"terraform/managed":   true,
		},
		CurrentTask: -1,
		Meta:        map[string]string{"feature-flags": "change-stage-v2", "field1": "value1", "field2": "value2"},
	}
	machine.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckMachineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				PreConfig: testAccCreateResources,
				Config:    testAccDrpMachine_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckMachineExists(t, "drp_machine.foo", &machine),
				),
			},
		},
	})
}

func testAccCreateResources() {
	config := testAccDrpProvider.Meta().(*Config)

	p := &models.Profile{Name: "p-test"}
	ta := &models.Param{Name: "terraform/allocated", Schema: map[string]string{"type": "boolean"}}
	tm := &models.Param{Name: "terraform/managed", Schema: map[string]string{"type": "boolean"}}
	tp := &models.Param{Name: "terraform/pool", Schema: map[string]string{"type": "string", "default": "default"}}
	m := &models.Machine{Name: "mach1", Secret: "12", Params: map[string]interface{}{"terraform/allocated": false, "terraform/managed": true}, Uuid: uuid.Parse("3945838b-be8c-4b35-8b1c-b538ddc71f7e")}

	config.session.CreateModel(p)
	config.session.CreateModel(ta)
	config.session.CreateModel(tm)
	config.session.CreateModel(tp)
	config.session.CreateModel(m)
}

func testAccDrpCheckMachineDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_machine" {
			continue
		}

		if _, err := config.session.GetModel("machines", rs.Primary.ID); err != nil {
			return fmt.Errorf("Machine does not exist")
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

		if found.Key() != rs.Primary.ID {
			return fmt.Errorf("Machine not found")
		}

		if err := diffObjects(machine, found, "Machine"); err != nil {
			return err
		}
		return nil
	}
}
