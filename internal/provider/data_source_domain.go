package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DomainDataSource{}

type DomainDataSource struct {
	client *Client
}

type DomainDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Expires              types.String `tfsdk:"expires"`
	DNSSEC               types.Bool   `tfsdk:"dnssec"`
	Autorenew            types.Bool   `tfsdk:"autorenew"`
	RenewalNotifications types.Bool   `tfsdk:"renewal_notifications"`
	NameserversCustom    types.Bool   `tfsdk:"nameservers_custom"`
	Delegated            types.String `tfsdk:"delegated"`
	HasPendingDNSSEC     types.Bool   `tfsdk:"has_pending_dnssec"`
	Reactivate           types.Bool   `tfsdk:"reactivate"`
	AuthKeyEnabled       types.Bool   `tfsdk:"auth_key_enabled"`
	SigningRequired      types.Bool   `tfsdk:"signing_required"`
}

func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

func (d *DomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *DomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a domain registered with Zone.EU.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the data source (same as name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The domain name.",
				Required:    true,
			},
			"expires": schema.StringAttribute{
				Description: "When the domain expires.",
				Computed:    true,
			},
			"dnssec": schema.BoolAttribute{
				Description: "Whether DNSSEC is enabled for the domain.",
				Computed:    true,
			},
			"autorenew": schema.BoolAttribute{
				Description: "Whether autorenew is enabled for the domain.",
				Computed:    true,
			},
			"renewal_notifications": schema.BoolAttribute{
				Description: "Whether renewal notification reminders are enabled.",
				Computed:    true,
			},
			"nameservers_custom": schema.BoolAttribute{
				Description: "Whether the domain uses custom nameservers.",
				Computed:    true,
			},
			"delegated": schema.StringAttribute{
				Description: "Username of the domain owner if the domain is delegated to you.",
				Computed:    true,
			},
			"has_pending_dnssec": schema.BoolAttribute{
				Description: "Whether the domain has a pending DNSSEC change.",
				Computed:    true,
			},
			"reactivate": schema.BoolAttribute{
				Description: "Whether the domain can be reactivated.",
				Computed:    true,
			},
			"auth_key_enabled": schema.BoolAttribute{
				Description: "Whether this TLD uses authorization key.",
				Computed:    true,
			},
			"signing_required": schema.BoolAttribute{
				Description: "Whether signing is required for the domain.",
				Computed:    true,
			},
		},
	}
}

func (d *DomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := d.client.GetDomain(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain",
			fmt.Sprintf("Could not read domain %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Also get preferences for renewal_notifications
	prefs, err := d.client.GetDomainPreferences(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain Preferences",
			fmt.Sprintf("Could not read domain preferences for %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	data.ID = data.Name
	data.Name = types.StringValue(domain.Name)
	data.Expires = types.StringValue(domain.Expires)
	data.DNSSEC = types.BoolValue(domain.DNSSEC)
	data.Autorenew = types.BoolValue(domain.Autorenew)
	data.RenewalNotifications = types.BoolValue(prefs.RenewalNotifications)
	data.NameserversCustom = types.BoolValue(domain.NameserversCustom)
	data.HasPendingDNSSEC = types.BoolValue(domain.HasPendingDNSSEC)
	data.Reactivate = types.BoolValue(domain.Reactivate)
	data.AuthKeyEnabled = types.BoolValue(domain.AuthKeyEnabled)
	data.SigningRequired = types.BoolValue(domain.SigningRequired)

	if domain.Delegated != "" {
		data.Delegated = types.StringValue(domain.Delegated)
	} else {
		data.Delegated = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
