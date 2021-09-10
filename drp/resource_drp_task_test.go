package drp

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpTask_basic = `
	resource "drp_task" "foo" {
		Name = "foo"
		Meta = {
			"feature-flags" = "sane-exit-codes"
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpTask_basic(t *testing.T) {
	task := models.Task{Name: "foo",
		Meta: map[string]string{"field1": "value1", "field2": "value2", "feature-flags": "sane-exit-codes"},
	}
	task.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDrpTask_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckTaskExists(t, "drp_task.foo", &task),
				),
			},
		},
	})
}

var testAccDrpTask_change_1 = `
	resource "drp_task" "foo" {
		Name = "foo"
		Description = "I am a task"
		Documentation = "I am docs"
		RequiredParams = [ "p1", "p2" ]
		OptionalParams = [ "p3", "p4" ]
		Templates = [
		  { Name = "t1", Path = "fred1", Contents = "temp1"},
		  { Name = "t2", Path = "fred2", Contents = "actual stuff"}
		]
	}`

var testAccDrpTask_change_2 = `
	resource "drp_task" "foo" {
		Name = "foo"
		Description = "I am a task again"
		Documentation = "I am docs more so"
		RequiredParams = [ "p3", "p4" ]
		OptionalParams = [ "p1", "p2" ]
		Templates = [
		  { Name = "t3", Path = "jill1", Contents = "temp2"},
		  { Name = "t4", Path = "jill2", Contents = "really actual stuff"}
		]
	}`

func TestAccDrpTask_change(t *testing.T) {
	task1 := models.Task{
		Name:           "foo",
		Description:    "I am a task",
		Documentation:  "I am docs",
		Meta:           map[string]string{"feature-flags": "sane-exit-codes"},
		RequiredParams: []string{"p1", "p2"},
		OptionalParams: []string{"p3", "p4"},
		Templates: []models.TemplateInfo{
			{Name: "t1", Path: "fred1", Contents: "temp1", Meta: map[string]string{}},
			{Name: "t2", Path: "fred2", Contents: "actual stuff", Meta: map[string]string{}},
		},
	}
	task1.Fill()
	task2 := models.Task{
		Name:           "foo",
		Description:    "I am a task again",
		Documentation:  "I am docs more so",
		Meta:           map[string]string{"feature-flags": "sane-exit-codes"},
		RequiredParams: []string{"p3", "p4"},
		OptionalParams: []string{"p1", "p2"},
		Templates: []models.TemplateInfo{
			{Name: "t3", Path: "jill1", Contents: "temp2", Meta: map[string]string{}},
			{Name: "t4", Path: "jill2", Contents: "really actual stuff", Meta: map[string]string{}},
		},
	}
	task2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDrpTask_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckTaskExists(t, "drp_task.foo", &task1),
				),
			},
			{
				Config: testAccDrpTask_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckTaskExists(t, "drp_task.foo", &task2),
				),
			},
		},
	})
}

func testAccDrpCheckTaskDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_task" {
			continue
		}

		if _, err := config.session.GetModel("tasks", rs.Primary.ID); err == nil {
			return fmt.Errorf("Task still exists")
		}
	}

	return nil
}

func testAccDrpCheckTaskExists(t *testing.T, n string, task *models.Task) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("tasks", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.Task)
		found.ClearValidation()

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("Task not found")
		}

		if err := diffObjects(task, found, "Task"); err != nil {
			return err
		}
		return nil
	}
}
