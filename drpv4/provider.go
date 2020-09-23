package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
)

/*
 * Enable terraform to use DRP as a provider.  Fill out the
 * appropriate functions and information about this plugin.
 */
func Provider() *schema.Provider {
	return &schema.Provider{

		ResourcesMap: map[string]*schema.Resource{
			"drp_machine": resourceMachine(),
		},

		// note yet, but potentially pools, params and profiles
		DataSourcesMap: map[string]*schema.Resource{},

		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Granted DRP token (use instead of RS_KEY)",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"RS_TOKEN",
				}, nil),
				ConflictsWith: []string{"key", "password"},
			},
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The DRP user:password key",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"RS_KEY",
				}, nil),
				ConflictsWith: []string{"token"},
			},
			"username": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The DRP user",
				ConflictsWith: []string{"key"},
			},
			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The DRP password",
				ConflictsWith: []string{"key", "token"},
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The DRP server URL. ie: https://1.2.3.4:8092",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"RS_ENDPOINT",
				}, nil),
			},
		},

		ConfigureContextFunc: providerConfigure,
	}
}

/*
 * The config method that terraform uses to pass information about configuration
 * to the plugin.
 */
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	log.Println("[DEBUG] Configuring the DRP provider")
	config := Config{
		endpoint: d.Get("endpoint").(string),
		username: d.Get("username").(string),
		password: d.Get("password").(string),
	}
	var diags diag.Diagnostics

	if token := d.Get("token"); token != nil {
		config.token = token.(string)
	}
	if key := d.Get("key"); key != "" {
		parts := strings.SplitN(key.(string), ":", 2)
		if len(parts) < 2 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Malformed DRP Credential",
				Detail:   fmt.Sprint("RS_KEY has not enough parts: ", key),
			})
			return nil, diags
		}
		config.username = parts[0]
		config.password = parts[1]
	}

	if config.token == "" && config.username == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing DRP Credential",
			Detail:   "drp provider requires username/password, credential, or token",
		})
		return nil, diags
	}
	if config.username != "" && config.password == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing DRP Password",
			Detail:   "drp provider requires a password for the specified user",
		})
		return nil, diags
	}

	log.Printf("[DEBUG] Attempting to connect with credentials %+v", config)
	if err := config.validateAndConnect(); err != nil {
		return nil, diag.FromErr(err)
	}

	info, err := config.session.Info()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to Connect",
			Detail:   fmt.Sprint("Failed to fetch info for ", config.endpoint),
		})
		return nil, diags
	}
	has_pool := false
	for _, f := range info.Features {
		if f == "embedded-pool" {
			has_pool = true
		}
	}
	if !has_pool {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Insufficient DRP Version",
			Detail:   fmt.Sprint("Pooling feature required.  Upgrade to v4.4 from ", info.Version),
		})
		return nil, diags
	}

	log.Printf("[Info] Digital Rebar %+v", info.Version)

	return &config, diags
}
