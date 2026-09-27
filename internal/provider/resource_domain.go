package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DomainResource{}
	_ resource.ResourceWithImportState = &DomainResource{}
)

type DomainResource struct {
	client *Client
}

type DomainResourceModel struct {
	Name                 types.String `tfsdk:"name"`
	Autorenew            types.Bool   `tfsdk:"autorenew"`
	DNSSEC               types.Bool   `tfsdk:"dnssec"`
	RenewalNotifications types.Bool   `tfsdk:"renewal_notifications"`
	NameserversCustom    types.Bool   `tfsdk:"nameservers_custom"`
	Expires              types.String `tfsdk:"expires"`
	Delegated            types.String `tfsdk:"delegated"`
	HasPendingDNSSEC     types.Bool   `tfsdk:"has_pending_dnssec"`
}

func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

func (r *DomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a domain's settings in Zone.EU. Note: This resource manages an existing domain - it does not register new domains.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The domain name. This is the unique identifier for the domain.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"autorenew": schema.BoolAttribute{
				Description: "Whether autorenew is enabled for the domain.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"dnssec": schema.BoolAttribute{
				Description: "Whether DNSSEC is enabled for the domain.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"renewal_notifications": schema.BoolAttribute{
				Description: "Whether renewal notification reminders are enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"nameservers_custom": schema.BoolAttribute{
				Description: "Whether the domain uses custom nameservers. Set to true to use custom nameservers, false to use Zone.EU default nameservers.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"expires": schema.StringAttribute{
				Description: "When the domain expires.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"delegated": schema.StringAttribute{
				Description: "Username of the domain owner if the domain is delegated to you.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"has_pending_dnssec": schema.BoolAttribute{
				Description: "Whether the domain has a pending DNSSEC change.",
				Computed:    true,
			},
		},
	}
}

func (r *DomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First verify the domain exists
	domain, err := r.client.GetDomain(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain",
			fmt.Sprintf("Could not read domain %s: %s. Note: This resource manages existing domains, it does not register new ones.", data.Name.ValueString(), err),
		)
		return
	}

	// Update domain settings
	autorenew := data.Autorenew.ValueBool()
	dnssec := data.DNSSEC.ValueBool()
	nsCustom := data.NameserversCustom.ValueBool()

	update := &DomainUpdate{
		Autorenew:         &autorenew,
		DNSSEC:            &dnssec,
		NameserversCustom: &nsCustom,
	}

	domain, err = r.client.UpdateDomain(data.Name.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Domain",
			fmt.Sprintf("Could not update domain %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Update preferences
	prefs := &DomainPreferences{
		RenewalNotifications: data.RenewalNotifications.ValueBool(),
	}
	_, err = r.client.UpdateDomainPreferences(data.Name.ValueString(), prefs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Domain Preferences",
			fmt.Sprintf("Could not update domain preferences for %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Set state
	data.Expires = types.StringValue(domain.Expires)
	data.HasPendingDNSSEC = types.BoolValue(domain.HasPendingDNSSEC)
	data.Autorenew = types.BoolValue(domain.Autorenew)
	data.DNSSEC = types.BoolValue(domain.DNSSEC)
	data.NameserversCustom = types.BoolValue(domain.NameserversCustom)
	data.RenewalNotifications = types.BoolValue(prefs.RenewalNotifications)

	if domain.Delegated != "" {
		data.Delegated = types.StringValue(domain.Delegated)
	} else {
		data.Delegated = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomain(data.Name.ValueString())
	if err != nil {
		// Handle 404 - domain no longer exists or is not accessible
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Domain",
			fmt.Sprintf("Could not read domain %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	prefs, err := r.client.GetDomainPreferences(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain Preferences",
			fmt.Sprintf("Could not read domain preferences for %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	data.Name = types.StringValue(domain.Name)
	data.Expires = types.StringValue(domain.Expires)
	data.DNSSEC = types.BoolValue(domain.DNSSEC)
	data.Autorenew = types.BoolValue(domain.Autorenew)
	data.RenewalNotifications = types.BoolValue(prefs.RenewalNotifications)
	data.NameserversCustom = types.BoolValue(domain.NameserversCustom)
	data.HasPendingDNSSEC = types.BoolValue(domain.HasPendingDNSSEC)

	if domain.Delegated != "" {
		data.Delegated = types.StringValue(domain.Delegated)
	} else {
		data.Delegated = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update domain settings
	autorenew := data.Autorenew.ValueBool()
	dnssec := data.DNSSEC.ValueBool()
	nsCustom := data.NameserversCustom.ValueBool()

	update := &DomainUpdate{
		Autorenew:         &autorenew,
		DNSSEC:            &dnssec,
		NameserversCustom: &nsCustom,
	}

	domain, err := r.client.UpdateDomain(data.Name.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Domain",
			fmt.Sprintf("Could not update domain %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Update preferences
	prefs := &DomainPreferences{
		RenewalNotifications: data.RenewalNotifications.ValueBool(),
	}
	_, err = r.client.UpdateDomainPreferences(data.Name.ValueString(), prefs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Domain Preferences",
			fmt.Sprintf("Could not update domain preferences for %s: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Set state
	data.Expires = types.StringValue(domain.Expires)
	data.HasPendingDNSSEC = types.BoolValue(domain.HasPendingDNSSEC)
	data.Autorenew = types.BoolValue(domain.Autorenew)
	data.DNSSEC = types.BoolValue(domain.DNSSEC)
	data.NameserversCustom = types.BoolValue(domain.NameserversCustom)
	data.RenewalNotifications = types.BoolValue(prefs.RenewalNotifications)

	if domain.Delegated != "" {
		data.Delegated = types.StringValue(domain.Delegated)
	} else {
		data.Delegated = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Domains cannot be deleted via API - we just remove from state
	// Optionally we could reset settings to defaults
	var data DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reset to safe defaults before removing from Terraform management
	autorenew := false
	dnssec := false

	update := &DomainUpdate{
		Autorenew: &autorenew,
		DNSSEC:    &dnssec,
	}

	_, _ = r.client.UpdateDomain(data.Name.ValueString(), update)
	// Ignore errors on delete - domain will just be removed from state
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by domain name
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
