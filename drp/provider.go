package drp

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var theResourcesMap = map[string]*schema.Resource{}
var theDataSourcesMap = map[string]*schema.Resource{}

/*
 * Enable terraform to use DRP as a provider.  Fill out the
 * appropriate functions and information about this plugin.
 */
func Provider() terraform.ResourceProvider {
	log.Println("[DEBUG] Initializing the DRP provider")
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The api key for API operations",
				DefaultFunc: schema.EnvDefaultFunc("RS_TOKEN", nil),
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
				DefaultFunc: schema.EnvDefaultFunc("RS_ENDPOINT", nil),
			},
		},

		ResourcesMap:   theResourcesMap,
		DataSourcesMap: theDataSourcesMap,

		ConfigureFunc: providerConfigure,
	}

	return p
}

func envDefaultKeyFunc(k, part string) schema.SchemaDefaultFunc {
	return func() (interface{}, error) {
		if v := os.Getenv(k); v != "" {
			parts := strings.SplitN(v, ":", 2)
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
	config := Config{
		Url: d.Get("api_url").(string),
	}

	if key := d.Get("api_key"); key != nil {
		config.Token = key.(string)
	}
	if user := d.Get("api_user"); user != nil {
		config.Username = user.(string)
		config.Password = d.Get("api_password").(string)
	}

	if config.Token == "" && config.Username == "" {
		return nil, fmt.Errorf("drp provider requires either user or token ids")
	}
	if config.Username != "" && config.Password == "" {
		return nil, fmt.Errorf("drp provider requires a password for the specified user")
	}

	if err := config.validateAndConnect(); err != nil {
		return nil, err
	}

	return &config, nil
}
