package drp

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/models"
)

var testAccDrpPlugin_basic = `
	resource "drp_plugin" "foo" {
		Name = "foo"
		PluginProvider = "ipmi"
		Meta = {
			"field1" = "value1"
			"field2" = "value2"
		}
	}`

func TestAccDrpPlugin_basic(t *testing.T) {
	plugin := models.Plugin{Name: "foo",
		Provider:     "ipmi",
		Meta:         map[string]string{"field1": "value1", "field2": "value2"},
		PluginErrors: []string{"Missing Plugin Provider: ipmi"},
	}
	plugin.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckPluginDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpPlugin_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckPluginExists(t, "drp_plugin.foo", &plugin),
				),
			},
		},
	})
}

var testAccDrpPlugin_change_1 = `
	resource "drp_plugin" "foo" {
		Name = "foo"
		PluginProvider = "ipmi"
		Description = "I am a plugin"
	}`

var testAccDrpPlugin_change_2 = `
	resource "drp_plugin" "foo" {
		Name = "foo"
		PluginProvider = "ipmi"
		Description = "I am a plugin again"
	}`

func TestAccDrpPlugin_change(t *testing.T) {
	plugin1 := models.Plugin{
		Name:         "foo",
		Description:  "I am a plugin",
		Provider:     "ipmi",
		PluginErrors: []string{"Missing Plugin Provider: ipmi"},
	}
	plugin1.Fill()
	plugin2 := models.Plugin{
		Name:         "foo",
		Description:  "I am a plugin again",
		Provider:     "ipmi",
		PluginErrors: []string{"Missing Plugin Provider: ipmi"},
	}
	plugin2.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckPluginDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpPlugin_change_1,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckPluginExists(t, "drp_plugin.foo", &plugin1),
				),
			},
			resource.TestStep{
				Config: testAccDrpPlugin_change_2,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckPluginExists(t, "drp_plugin.foo", &plugin2),
				),
			},
		},
	})
}

var testAccDrpPlugin_withParams = `
	resource "drp_plugin" "foo" {
		Name = "foo"
		PluginProvider = "ipmi"
		Params = {
			"test/string" = "fred"
			"test/int" = 3
			"test/bool" = "true"
			"test/list" = "[\"one\",\"two\"]"
		}
	}`

func TestAccDrpPlugin_withParams(t *testing.T) {
	plugin := models.Plugin{Name: "foo", Provider: "ipmi",
		Params: map[string]interface{}{
			"test/string": "fred",
			"test/int":    3,
			"test/bool":   true,
			"test/list":   []string{"one", "two"},
		},
		PluginErrors: []string{"Missing Plugin Provider: ipmi"},
	}
	plugin.Fill()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccDrpPreCheck(t) },
		Providers:    testAccDrpProviders,
		CheckDestroy: testAccDrpCheckPluginDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDrpPlugin_withParams,
				Check: resource.ComposeTestCheckFunc(
					testAccDrpCheckPluginExists(t, "drp_plugin.foo", &plugin),
				),
			},
		},
	})
}

func testAccDrpCheckPluginDestroy(s *terraform.State) error {
	config := testAccDrpProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "drp_plugin" {
			continue
		}

		if _, err := config.session.GetModel("plugins", rs.Primary.ID); err == nil {
			return fmt.Errorf("Plugin still exists")
		}
	}

	return nil
}

func testAccDrpCheckPluginExists(t *testing.T, n string, plugin *models.Plugin) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccDrpProvider.Meta().(*Config)

		obj, err := config.session.GetModel("plugins", rs.Primary.ID)
		if err != nil {
			return err
		}
		found := obj.(*models.Plugin)
		found.ClearValidation()

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("Plugin not found")
		}

		if err := diffObjects(plugin, found, "Plugin"); err != nil {
			return err
		}
		return nil
	}
}
