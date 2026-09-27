package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	envVarUsername         = "ZONE_EU_USERNAME"
	envVarAPIKey           = "ZONE_EU_API_KEY"
	errMissingAuthUsername = "Required username could not be found. Please set the username using an input variable in the provider configuration block or by using the `" + envVarUsername + "` environment variable."
	errMissingAuthAPIKey   = "Required api_key could not be found. Please set the api_key using an input variable in the provider configuration block or by using the `" + envVarAPIKey + "` environment variable."
)

var _ provider.Provider = &ZoneProvider{}

type ZoneProvider struct {
	version string
}

type ZoneProviderModel struct {
	Username types.String `tfsdk:"username"`
	APIKey   types.String `tfsdk:"api_key"`
}

func (p *ZoneProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zoneeu"
	resp.Version = p.version
}

func (p *ZoneProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Zone.EU provider is used to manage DNS records and domains on Zone.EU hosting platform.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "The ZoneID username used to authenticate with Zone.EU API. Can also be set via the ZONE_EU_USERNAME environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key used to authenticate with Zone.EU API. Can also be set via the ZONE_EU_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *ZoneProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ZoneProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get username from config or environment
	username := config.Username.ValueString()
	if username == "" {
		username = os.Getenv(envVarUsername)
	}
	if username == "" {
		resp.Diagnostics.AddError("Missing Username", errMissingAuthUsername)
		return
	}

	// Get API key from config or environment
	apiKey := config.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv(envVarAPIKey)
	}
	if apiKey == "" {
		resp.Diagnostics.AddError("Missing API Key", errMissingAuthAPIKey)
		return
	}

	client := NewClient(username, apiKey)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ZoneProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSARecordResource,
		NewDNSAAAARecordResource,
		NewDNSCNAMERecordResource,
		NewDNSMXRecordResource,
		NewDNSTXTRecordResource,
		NewDNSNSRecordResource,
		NewDNSSRVRecordResource,
		NewDNSCAARecordResource,
		NewDNSTLSARecordResource,
		NewDNSSSHFPRecordResource,
		NewDNSURLRecordResource,
		NewDomainResource,
		NewDomainNameserverResource,
	}
}

func (p *ZoneProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDNSZoneDataSource,
		NewDomainDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ZoneProvider{
			version: version,
		}
	}
}

