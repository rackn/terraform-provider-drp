package drp

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpUser_basic = `
	resource "drp_user" "foo" {
		Name = "foo"
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpUser_basic(t *testing.T) {
	user := models.User{Name: "foo",
		Meta: map[string]string{"field1": "value1", "field2": "value2"},
	}
	user.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckUserDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpUser_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckUserExists(t, "drp_user.foo", &user),
				),
			},
		},
	})
}

var testAccDrpUser_change_1 = `
	resource "drp_user" "foo" {
		Name = "foo"
		Secret = "I am a user"
	}`

var testAccDrpUser_change_2 = `
	resource "drp_user" "foo" {
		Name = "foo"
		Secret = "I am a user again"
	}`

func TestAccDrpUser_change(t *testing.T) {
	user1 := models.User{Name: "foo", Secret: "I am a user"}
	user1.Fill()
	user2 := models.User{Name: "foo", Secret: "I am a user again"}
	user2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckUserDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpUser_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckUserExists(t, "drp_user.foo", &user1),
				),
			},
			resource.TestStep{
				Config: testAccDrpUser_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckUserExists(t, "drp_user.foo", &user2),
				),
			},
		},
	})
}

// XXX: One day worry about setting user's passwords.

func testAccDrpCheckUserDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_user" {
			continue
		}

		if _, err := config.session.GetModel("users", rs.Primary.ID); err == nil {
			return fmt.Errorf("User still exists")
		}
	}

	return nil
}

func testAccDrpCheckUserExists(t *testing.T, n string, user *models.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("users", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.User)
		found.ClearValidation()

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("User not found")
		}

		// Secret is unset, it should be set to something
		if user.Secret == "" {
			if found.Secret != "" {
				user.Secret = found.Secret
			}
		}

		if err := diffObjects(user, found, "User"); err != nil {
			return err
		}
		return nil
	}
}
