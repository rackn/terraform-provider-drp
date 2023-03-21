package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.com/rackn/provision/v4/api"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &Config{}

type ConfigModel struct {
	Token    types.String `tfsdk:"token"`
	Key      types.String `tfsdk:"key"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Endpoint types.String `tfsdk:"endpoint"`
}

type Config struct {
	token    string
	username string
	password string
	endpoint string
	session  *api.Client
	version  string
}

/*
 * Builds a client object for this config
 */
func (c *Config) validateAndConnect(ctx context.Context) error {
	tflog.Debug(ctx, "[Config.validateAndConnect] Configuring the DRP API client")

	if c.session != nil {
		return nil
	}
	var err error
	if c.token != "" {
		c.session, err = api.TokenSession(c.endpoint, c.token)
	} else {
		c.session, err = api.UserSession(c.endpoint, c.username, c.password)
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating session: %s", err.Error()))
		return fmt.Errorf("Error creating session: %s", err)
	} else {
		tflog.Debug(ctx, fmt.Sprintf("[Condig.validateAndConnect] Authenticated! Session Starting w %+v", c))
	}

	return nil
}

func (p *Config) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "drp"
	resp.Version = p.version
}

func (p *Config) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:            true,
				Description:         "Granted DRP token (use instead of RS_KEY)",
				MarkdownDescription: "Granted DRP token (use instead of RS_KEY)",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password")),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("key")),
				},
			},
			"key": schema.StringAttribute{
				Optional:            true,
				Description:         "The DRP user:password key",
				MarkdownDescription: "The DRP user:password key",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("token")),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("username")),
				},
			},
			"username": schema.StringAttribute{
				Optional:            true,
				Description:         "The DRP user",
				MarkdownDescription: "The DRP user",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("token")),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("key")),
				},
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Description:         "The DRP password",
				MarkdownDescription: "The DRP password",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("token")),
				},
			},
			"endpoint": schema.StringAttribute{
				Required:            true,
				Description:         "The DRP server URL. ie: https://1.2.3.4:8092",
				MarkdownDescription: "The DRP server URL. ie: https://1.2.3.4:8092",
			},
		},
	}
}

/*
 * The config method that terraform uses to pass information about configuration
 * to the plugin.
 */
func (p *Config) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	p.endpoint = os.Getenv("RS_ENDPOINT")
	p.username = os.Getenv("RS_USERNAME")
	p.password = os.Getenv("RS_PASSWORD")
	p.token = os.Getenv("RS_TOKEN")
	key := os.Getenv("RS_KEY")
	if key != "" {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) < 2 {
			resp.Diagnostics.AddError("Malformed RS_KEY", "While configuring the provider, the RS_KEY environment variable is malformed.")
			return
		}
		p.username = parts[0]
		p.password = parts[1]
	}
	var data ConfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if token := data.Token.ValueString(); token != "" {
		p.token = token
	}
	if key := data.Key.ValueString(); key != "" {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) < 2 {
			resp.Diagnostics.AddError("Malformed key", "While configuring the provider, the key attribute is malformed.")
			return
		}
		p.username = parts[0]
		p.password = parts[1]
	}
	if username := data.Username.ValueString(); username != "" {
		p.username = username
	}
	if password := data.Password.ValueString(); password != "" {
		p.password = password
	}
	if endpoint := data.Endpoint.ValueString(); endpoint != "" {
		p.endpoint = endpoint
	}

	if p.token == "" && p.username == "" {
		resp.Diagnostics.AddError("Malformed DRP credentials", "While configuring the provider, the key, token or username/password attributes must be provided.")
		return
	}
	if p.username != "" && p.password == "" {
		resp.Diagnostics.AddError("Missing DRP password", "While configuring the provider, the password attribute was not specified.")
		return
	}

	log.Printf("[DEBUG] Attempting to connect with credentials %+v", *p)
	if err := p.validateAndConnect(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to create DRP client", err.Error())
		return
	}

	info, err := p.session.Info()
	if err != nil {
		resp.Diagnostics.AddError("Failed to Connect", fmt.Sprint("Failed to fetch info for ", p.endpoint))
		return
	}
	has_pool := false
	for _, f := range info.Features {
		if f == "embedded-pool" {
			has_pool = true
		}
	}
	if !has_pool {
		resp.Diagnostics.AddError("Insufficient DRP Version", fmt.Sprint("Pooling feature required.  Upgrade to v4.4 from ", info.Version))
	}

	log.Printf("[Info] Digital Rebar %+v", info.Version)
	resp.ResourceData = p
}

func (p *Config) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMachineResource,
	}
}

func (p *Config) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Config{
			version: version,
		}
	}
}
