package drp

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

/*
 * Enable terraform to use DRP as a provider.  Fill out the
 * appropriate functions and information about this plugin.
 */
func Provider() terraform.ResourceProvider {
	log.Println("[DEBUG] Initializing the DRP provider")
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The api key for API operations",
				DefaultFunc: envDefaultFunc("RS_TOKEN"),
			},
			"api_user": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The api user for API operations",
				DefaultFunc: envDefaultKeyFunc("RS_KEY", "username"),
			},
			"api_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The api password for API operations",
				DefaultFunc: envDefaultKeyFunc("RS_KEY", "password"),
			},
			"api_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The DRP server URL. ie: https://1.2.3.4:8092",
				DefaultFunc: envDefaultFunc("RS_ENDPOINT"),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"drp_machine": resourceDRPMachine(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func envDefaultFunc(k string) schema.SchemaDefaultFunc {
	return func() (interface{}, error) {
		if v := os.Getenv(k); v != "" {
			return v, nil
		}

		return nil, nil
	}
}

func envDefaultKeyFunc(k, part string) schema.SchemaDefaultFunc {
	return func() (interface{}, error) {
		if v := os.Getenv(k); v != "" {
			parts := strings.SplitN(kv, ":", 2)
			if len(parts) < 2 {
				return nil, fmt.Errorf("RS_KEY has not enough parts")
			}
			if part == "username" {
				return parts[0], nil
			} else if part == "password" {
				return parts[1], nil
			}
			return nil, fmt.Errorf("Asking for unknown part of RS_KEY: %s", part)
		}

		return nil, nil
	}
}

/*
 * The config method that terraform uses to pass information about configuration
 * to the plugin.
 */
func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	log.Println("[DEBUG] Configuring the DRP provider")
	cc := Client{
		APIURL: d.Get("api_url").(string),
	}

	if key := d.Get("api_key"); key != nil {
		cc.APIKey = key.(string)
	}
	if user := d.Get("api_user"); user != nil {
		cc.APIUser = user.(string)
		cc.APIPassword = d.Get("api_password").(string)
	}

	if cc.APIKey == "" && cc.APIUser == "" {
		return nil, fmt.Errorf("drp provider requires either user or token ids")
	}
	if cc.APIUser != "" && cc.APIPassword == "" {
		return nil, fmt.Errorf("drp provider requires a password for the specified user")
	}

	return cc.Client()
}
