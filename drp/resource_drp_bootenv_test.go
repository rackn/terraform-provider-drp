package drp

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpBootEnv_basic = `
	resource "drp_bootenv" "foo" {
		Name = "foo"
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
		OS = {}
	}`

func TestAccDrpBootEnv_basic(t *testing.T) {
	bootenv := models.BootEnv{Name: "foo",
		Meta: map[string]string{"field1": "value1", "field2": "value2"},
	}
	bootenv.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckBootEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDrpBootEnv_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckBootEnvExists(t, "drp_bootenv.foo", &bootenv),
				),
			},
		},
	})
}

var testAccDrpBootEnv_change_1 = `
	resource "drp_bootenv" "foo" {
		Name = "foo"
		Description = "I am a bootenv"
		RequiredParams = [ "p1", "p2" ]
		OptionalParams = [ "p3", "p4" ]
		Templates = [
		  { Name = "t1", Path = "fred1", Contents = "temp1"},
		  { Name = "t2", Path = "fred2", Contents = "actual stuff"}
		]
		Kernel = "jill"
		Initrds = [ "joyce", "julie", "janna" ]
		BootParams = "kernel go"
		OnlyUnknown = true
		OS = {}
	}`

var testAccDrpBootEnv_change_2 = `
	resource "drp_bootenv" "foo" {
		Name = "foo"
		Description = "I am a bootenv again"
		RequiredParams = [ "p3", "p4" ]
		OptionalParams = [ "p1", "p2" ]
		Templates = [
		  { Name = "t3", Path = "jill1", Contents = "temp2"},
		  { Name = "t4", Path = "jill2", Contents = "really actual stuff"}
		]
		Kernel = "~jill"
		Initrds = [ "~joyce", "~julie", "~janna" ]
		BootParams = "kernel nogo"
		OnlyUnknown = false
		OS = {}
	}`

func TestAccDrpBootEnv_change(t *testing.T) {
	bootenv1 := models.BootEnv{
		Name:           "foo",
		Description:    "I am a bootenv",
		RequiredParams: []string{"p1", "p2"},
		OptionalParams: []string{"p3", "p4"},
		Templates: []models.TemplateInfo{
			{Name: "t1", Path: "fred1", Contents: "temp1", Meta: map[string]string{}},
			{Name: "t2", Path: "fred2", Contents: "actual stuff", Meta: map[string]string{}},
		},
		Kernel:      "jill",
		Initrds:     []string{"joyce", "julie", "janna"},
		BootParams:  "kernel go",
		OnlyUnknown: true,
	}
	bootenv1.Fill()
	bootenv2 := models.BootEnv{
		Name:           "foo",
		Description:    "I am a bootenv again",
		RequiredParams: []string{"p3", "p4"},
		OptionalParams: []string{"p1", "p2"},
		Templates: []models.TemplateInfo{
			{Name: "t3", Path: "jill1", Contents: "temp2", Meta: map[string]string{}},
			{Name: "t4", Path: "jill2", Contents: "really actual stuff", Meta: map[string]string{}},
		},
		Kernel:     "~jill",
		Initrds:    []string{"~joyce", "~julie", "~janna"},
		BootParams: "kernel nogo",
	}
	bootenv2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckBootEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDrpBootEnv_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckBootEnvExists(t, "drp_bootenv.foo", &bootenv1),
				),
			},
			{
				Config: testAccDrpBootEnv_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckBootEnvExists(t, "drp_bootenv.foo", &bootenv2),
				),
			},
		},
	})
}

func testAccDrpCheckBootEnvDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_bootenv" {
			continue
		}

		if _, err := config.session.GetModel("bootenvs", rs.Primary.ID); err == nil {
			return fmt.Errorf("BootEnv still exists")
		}
	}

	return nil
}

func testAccDrpCheckBootEnvExists(t *testing.T, n string, bootenv *models.BootEnv) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("bootenvs", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.BootEnv)
		found.ClearValidation()

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("BootEnv not found")
		}

		if err := diffObjects(bootenv, found, "BootEnv"); err != nil {
			return err
		}
		return nil
	}
}
