package drp

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/digitalrebar/provision/models"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var testAccDrpParam_basic = `
	resource "drp_param" "foo" {
		Name = "foo"
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpParam_basic(t *testing.T) {
	param := models.Param{Name: "foo",
		Meta: map[string]string{"field1": "value1", "field2": "value2"},
	}
	param.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckParamDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpParam_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckParamExists(t, "drp_param.foo", &param),
				),
			},
		},
	})
}

var testAccDrpParam_change_1 = `
	resource "drp_param" "foo" {
		Name = "foo"
		Description = "I am a param"
		Documentation = "here I am"
		Schema = {
			"type" = "boolean"
		}
	}`

var testAccDrpParam_change_2 = `
	resource "drp_param" "foo" {
		Name = "foo"
		Description = "I am a param again"
		Documentation = "here am I"
		Schema = {
			"type" = "int"
		}
	}`

func TestAccDrpParam_change(t *testing.T) {
	param1 := models.Param{Name: "foo", Description: "I am a param", Documentation: "here I am"}
	param1.Fill()
	param2 := models.Param{Name: "foo", Description: "I am a param again", Documentation: "here am I"}
	param2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckParamDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpParam_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckParamExists(t, "drp_param.foo", &param1),
				),
			},
			resource.TestStep{
				Config: testAccDrpParam_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckParamExists(t, "drp_param.foo", &param2),
				),
			},
		},
	})
}

func testAccDrpCheckParamDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_param" {
			continue
		}

		if _, err := config.session.GetModel("params", rs.Primary.ID); err == nil {
			return fmt.Errorf("Param still exists")
		}
	}

	return nil
}

func testAccDrpCheckParamExists(t *testing.T, n string, param *models.Param) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("params", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.Param)
		found.ClearValidation()

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("Param not found")
		}

		if !reflect.DeepEqual(param, found) {
			return fmt.Errorf("Param doesn't match: e:%#v a:%#v", param, found)
		}
		return nil
	}
}
