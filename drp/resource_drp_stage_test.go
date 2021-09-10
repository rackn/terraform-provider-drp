package drp

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpStage_basic = `
	resource "drp_stage" "foo" {
		Name = "foo"
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpStage_basic(t *testing.T) {
	stage := models.Stage{Name: "foo",
		Meta:       map[string]string{"field1": "value1", "field2": "value2"},
		RunnerWait: true,
	}
	stage.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDrpStage_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckStageExists(t, "drp_stage.foo", &stage),
				),
			},
		},
	})
}

var testAccDrpStage_change_1 = `
	resource "drp_stage" "foo" {
		Name = "foo"
		Description = "I am a stage"
		RequiredParams = [ "p1", "p2" ]
		OptionalParams = [ "p3", "p4" ]
		Templates = [
		  { Name = "t1", Path = "fred1", Contents = "temp1"},
		  { Name = "t2", Path = "fred2", Contents = "actual stuff"}
		]
		BootEnv = "local"
		Tasks = [ "t1", "t2" ]
		Profiles = [ "p1" ]
		Reboot = true
		RunnerWait = true
	}`

var testAccDrpStage_change_2 = `
	resource "drp_stage" "foo" {
		Name = "foo"
		Description = "I am a stage again"
		RequiredParams = [ "p3", "p4" ]
		OptionalParams = [ "p1", "p2" ]
		Templates = [
		  { Name = "t3", Path = "jill1", Contents = "temp2"},
		  { Name = "t4", Path = "jill2", Contents = "really actual stuff"}
		]
		Tasks = [ "t3", "t4", "t5" ]
		Profiles = [ "p2", "p3" ]
		Reboot = false
		RunnerWait = true
	}`

func TestAccDrpStage_change(t *testing.T) {
	stage1 := models.Stage{
		Name:           "foo",
		Description:    "I am a stage",
		RequiredParams: []string{"p1", "p2"},
		OptionalParams: []string{"p3", "p4"},
		Templates: []models.TemplateInfo{
			{Name: "t1", Path: "fred1", Contents: "temp1", Meta: map[string]string{}},
			{Name: "t2", Path: "fred2", Contents: "actual stuff", Meta: map[string]string{}},
		},
		BootEnv:    "local",
		Tasks:      []string{"t1", "t2"},
		Profiles:   []string{"p1"},
		Reboot:     true,
		RunnerWait: true,
	}
	stage1.Fill()
	stage2 := models.Stage{
		Name:           "foo",
		Description:    "I am a stage again",
		RequiredParams: []string{"p3", "p4"},
		OptionalParams: []string{"p1", "p2"},
		Templates: []models.TemplateInfo{
			{Name: "t3", Path: "jill1", Contents: "temp2", Meta: map[string]string{}},
			{Name: "t4", Path: "jill2", Contents: "really actual stuff", Meta: map[string]string{}},
		},
		BootEnv:    "local",
		Tasks:      []string{"t3", "t4", "t5"},
		Profiles:   []string{"p2", "p3"},
		RunnerWait: true,
	}
	stage2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDrpStage_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckStageExists(t, "drp_stage.foo", &stage1),
				),
			},
			{
				Config: testAccDrpStage_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckStageExists(t, "drp_stage.foo", &stage2),
				),
			},
		},
	})
}

func testAccDrpCheckStageDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_stage" {
			continue
		}

		if _, err := config.session.GetModel("stages", rs.Primary.ID); err == nil {
			return fmt.Errorf("Stage still exists")
		}
	}

	return nil
}

func testAccDrpCheckStageExists(t *testing.T, n string, stage *models.Stage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("stages", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.Stage)
		found.ClearValidation()

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("Stage not found")
		}

		if err := diffObjects(stage, found, "Stage"); err != nil {
			return err
		}
		return nil
	}
}
