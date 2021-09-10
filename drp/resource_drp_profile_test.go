package drp

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpProfile_basic = `
	resource "drp_profile" "foo" {
		Name = "foo"
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpProfile_basic(t *testing.T) {
	profile := models.Profile{Name: "foo",
		Meta: map[string]string{"field1": "value1", "field2": "value2"},
	}
	profile.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckProfileDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpProfile_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckProfileExists(t, "drp_profile.foo", &profile),
				),
			},
		},
	})
}

var testAccDrpProfile_change_1 = `
	resource "drp_profile" "foo" {
		Name = "foo"
		Description = "I am a profile"
	}`

var testAccDrpProfile_change_2 = `
	resource "drp_profile" "foo" {
		Name = "foo"
		Description = "I am a profile again"
	}`

func TestAccDrpProfile_change(t *testing.T) {
	profile1 := models.Profile{Name: "foo", Description: "I am a profile"}
	profile1.Fill()
	profile2 := models.Profile{Name: "foo", Description: "I am a profile again"}
	profile2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckProfileDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpProfile_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckProfileExists(t, "drp_profile.foo", &profile1),
				),
			},
			resource.TestStep{
				Config: testAccDrpProfile_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckProfileExists(t, "drp_profile.foo", &profile2),
				),
			},
		},
	})
}

var testAccDrpProfile_withParams = `
	resource "drp_profile" "foo" {
		Name = "foo"
		Description = "I am a profile again"
		Params = {
			"test/string" = "fred"
			"test/int" = 3
			"test/bool" = "true"
			"test/list" = "[\"one\",\"two\"]"
		}
	}`

func TestAccDrpProfile_withParams(t *testing.T) {
	profile := models.Profile{Name: "foo", Description: "I am a profile again",
		Params: map[string]interface{}{
			"test/string": "fred",
			"test/int":    3,
			"test/bool":   true,
			"test/list":   []string{"one", "two"},
		},
	}
	profile.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckProfileDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpProfile_withParams,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckProfileExists(t, "drp_profile.foo", &profile),
				),
			},
		},
	})
}

func testAccDrpCheckProfileDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_profile" {
			continue
		}

		if _, err := config.session.GetModel("profiles", rs.Primary.ID); err == nil {
			return fmt.Errorf("Profile still exists")
		}
	}

	return nil
}

func testAccDrpCheckProfileExists(t *testing.T, n string, profile *models.Profile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("profiles", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.Profile)
		found.ClearValidation()

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("Profile not found")
		}

		if err := diffObjects(profile, found, "Profile"); err != nil {
			return err
		}
		return nil
	}
}
