package drp

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpTemplate_basic = `
	resource "drp_template" "foo" {
		ID = "foo"
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpTemplate_basic(t *testing.T) {
	template := models.Template{ID: "foo",
		Meta: map[string]string{"field1": "value1", "field2": "value2"},
	}
	template.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckTemplateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpTemplate_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckTemplateExists(t, "drp_template.foo", &template),
				),
			},
		},
	})
}

var testAccDrpTemplate_change_1 = `
	resource "drp_template" "foo" {
		ID = "foo"
		Description = "I am a template"
		Contents = "base content"
	}`

var testAccDrpTemplate_change_2 = `
	resource "drp_template" "foo" {
		ID = "foo"
		Description = "I am a template again"
		Contents = "{{ .Env.OS }}"
	}`

func TestAccDrpTemplate_change(t *testing.T) {
	template1 := models.Template{ID: "foo", Description: "I am a template", Contents: "base content"}
	template1.Fill()
	template2 := models.Template{ID: "foo", Description: "I am a template again", Contents: "{{ .Env.OS }}"}
	template2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckTemplateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpTemplate_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckTemplateExists(t, "drp_template.foo", &template1),
				),
			},
			resource.TestStep{
				Config: testAccDrpTemplate_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckTemplateExists(t, "drp_template.foo", &template2),
				),
			},
		},
	})
}

func testAccDrpCheckTemplateDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_template" {
			continue
		}

		if _, err := config.session.GetModel("templates", rs.Primary.ID); err == nil {
			return fmt.Errorf("Template still exists")
		}
	}

	return nil
}

func testAccDrpCheckTemplateExists(t *testing.T, n string, template *models.Template) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("templates", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.Template)
		found.ClearValidation()

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Template not found")
		}

		if err := diffObjects(template, found, "Template"); err != nil {
			return err
		}
		return nil
	}
}
