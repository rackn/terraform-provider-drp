package provider

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/rackn/terraform-provider-drp/client"
)

/*
 * Enable terraform to use us as a provider.  Fill out the
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
			},
			"api_user": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The api user for API operations",
			},
			"api_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The api password for API operations",
			},
			"api_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The DRP server URL. ie: https://1.2.3.4:8092",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"drp_machine": resourceDRPMachine(),
		},

		ConfigureFunc: providerConfigure,
	}
}

/*
 * The config method that terraform uses to pass information about configuration
 * to the plugin.
 */
func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	log.Println("[DEBUG] Configuring the DRP provider")
	cc := client.Client{
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
