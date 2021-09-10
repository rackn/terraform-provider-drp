package drp

import (
	"fmt"
	"net"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pborman/uuid"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpRawMachine_basic = `
	resource "drp_raw_machine" "foo" {
		Name = "mach11"
		Uuid = "3945838b-be8c-4b35-8b1c-b538ddc71f7c"
		Secret = "12"
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
			"feature-flags" = "change-stage-v2"
		}
	}`

func TestAccDrpRawMachine_basic(t *testing.T) {
	raw_machine := models.Machine{Name: "mach11",
		Uuid:        uuid.Parse("3945838b-be8c-4b35-8b1c-b538ddc71f7c"),
		Secret:      "12",
		Stage:       "none",
		BootEnv:     "local",
		Runnable:    true,
		CurrentTask: -1,
		Meta:        map[string]string{"feature-flags": "change-stage-v2", "field1": "value1", "field2": "value2"},
	}
	raw_machine.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckRawMachineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpRawMachine_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckRawMachineExists(t, "drp_raw_machine.foo", &raw_machine),
				),
			},
		},
	})
}

var testAccDrpRawMachine_change_1 = `
        resource "drp_profile" "p1" {
		Name = "p1"
	}
        resource "drp_profile" "p2" {
		Name = "p2"
	}
        resource "drp_task" "t1" {
		Name = "t1"
	}
        resource "drp_task" "t2" {
		Name = "t2"
	}
	resource "drp_raw_machine" "foo" {
		depends_on = ["drp_profile.p1", "drp_profile.p2", "drp_task.t1", "drp_task.t2"]
		Name = "mach11"
		Uuid = "3945838b-be8c-4b35-8b1c-b538ddc71f7c"
		Secret = "12"
		Description = "I am a raw_machine"
		CurrentJob = "3945838b-be8c-4b35-8b1c-b538ddc71f7f"
		Address = "1.1.1.1"
		Stage = "none"
		BootEnv = "local"
		Profiles = [ "p1", "p2" ]
		Tasks = [ "t1", "t2" ]
		Runnable = true
		OS = "fred"
	}`

var testAccDrpRawMachine_change_2 = `
        resource "drp_profile" "p1" {
		Name = "p1"
	}
        resource "drp_profile" "p2" {
		Name = "p2"
	}
        resource "drp_task" "t1" {
		Name = "t1"
	}
        resource "drp_task" "t2" {
		Name = "t2"
	}
        resource "drp_profile" "p3" {
		Name = "p3"
	}
        resource "drp_profile" "p4" {
		Name = "p4"
	}
        resource "drp_task" "t3" {
		Name = "t3"
	}
        resource "drp_task" "t4" {
		Name = "t4"
	}
	resource "drp_raw_machine" "foo" {
		depends_on = ["drp_profile.p3", "drp_profile.p4", "drp_task.t3", "drp_task.t4", "drp_profile.p1", "drp_profile.p2", "drp_task.t1", "drp_task.t2"]
		Name = "mach11"
		Uuid = "3945838b-be8c-4b35-8b1c-b538ddc71f7c"
		Secret = "12"
		Description = "I am a raw_machine again"
		CurrentJob = "3945838b-be8c-4b35-8b1c-b538ddc71f7a"
		Address = "1.1.1.2"
		Stage = "none"
		BootEnv = "local"
		Profiles = [ "p3", "p4" ]
		Tasks = [ "t3", "t4" ]
		Runnable = false
		OS = "greg"
	}`

func TestAccDrpRawMachine_change(t *testing.T) {
	raw_machine1 := models.Machine{
		Name:        "mach11",
		Address:     net.ParseIP("1.1.1.1"),
		Description: "I am a raw_machine",
		Uuid:        uuid.Parse("3945838b-be8c-4b35-8b1c-b538ddc71f7c"),
		CurrentJob:  uuid.Parse("3945838b-be8c-4b35-8b1c-b538ddc71f7f"),
		CurrentTask: -1,
		Secret:      "12",
		Stage:       "none",
		BootEnv:     "local",
		Runnable:    true,
		Profiles:    []string{"p1", "p2"},
		Tasks:       []string{"t1", "t2"},
		OS:          "fred",
		Meta:        map[string]string{"feature-flags": "change-stage-v2"},
	}
	raw_machine1.Fill()
	raw_machine2 := models.Machine{
		Name:        "mach11",
		Address:     net.ParseIP("1.1.1.2"),
		Description: "I am a raw_machine again",
		Uuid:        uuid.Parse("3945838b-be8c-4b35-8b1c-b538ddc71f7c"),
		CurrentJob:  uuid.Parse("3945838b-be8c-4b35-8b1c-b538ddc71f7a"),
		CurrentTask: -1,
		Secret:      "12",
		Stage:       "none",
		BootEnv:     "local",
		Profiles:    []string{"p3", "p4"},
		Tasks:       []string{"t3", "t4"},
		Runnable:    false,
		OS:          "greg",
		Meta:        map[string]string{"feature-flags": "change-stage-v2"},
	}
	raw_machine2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckRawMachineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpRawMachine_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckRawMachineExists(t, "drp_raw_machine.foo", &raw_machine1),
				),
			},
			resource.TestStep{
				Config: testAccDrpRawMachine_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckRawMachineExists(t, "drp_raw_machine.foo", &raw_machine2),
				),
			},
		},
	})
}

var testAccDrpRawMachine_withParams = `
	resource "drp_raw_machine" "foo" {
		Name = "mach11"
		Uuid = "3945838b-be8c-4b35-8b1c-b538ddc71f7c"
		Secret = "12"
		Params = {
			"test/string" = "fred"
			"test/int" = "3"
			"test/bool" = "true"
			"test/list" = "[\"one\",\"two\"]"
		}
	}`

func TestAccDrpRawMachine_withParams(t *testing.T) {
	raw_machine := models.Machine{
		Name:        "mach11",
		Uuid:        uuid.Parse("3945838b-be8c-4b35-8b1c-b538ddc71f7c"),
		Secret:      "12",
		Stage:       "none",
		BootEnv:     "local",
		Runnable:    true,
		CurrentTask: -1,
		Meta:        map[string]string{"feature-flags": "change-stage-v2"},
		Params: map[string]interface{}{
			"test/string": "fred",
			"test/int":    3,
			"test/bool":   true,
			"test/list":   []string{"one", "two"},
		},
	}
	raw_machine.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckRawMachineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpRawMachine_withParams,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckRawMachineExists(t, "drp_raw_machine.foo", &raw_machine),
				),
			},
		},
	})
}

func testAccDrpCheckRawMachineDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_raw_machine" {
			continue
		}

		if _, err := config.session.GetModel("raw_machines", rs.Primary.ID); err == nil {
			return fmt.Errorf("RawMachine still exists")
		}
	}

	return nil
}

func testAccDrpCheckRawMachineExists(t *testing.T, n string, raw_machine *models.Machine) resource.TestCheckFunc {
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

		if found.Uuid.String() != rs.Primary.ID {
			return fmt.Errorf("RawMachine not found")
		}

		if err := diffObjects(raw_machine, found, "RawMachine"); err != nil {
			return err
		}
		return nil
	}
}
