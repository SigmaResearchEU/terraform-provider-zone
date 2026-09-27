package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DomainNameserverResource{}
	_ resource.ResourceWithImportState = &DomainNameserverResource{}
)

type DomainNameserverResource struct {
	client *Client
}

type DomainNameserverResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Domain   types.String   `tfsdk:"domain"`
	Hostname types.String   `tfsdk:"hostname"`
	IP       []types.String `tfsdk:"ip"`
}

func NewDomainNameserverResource() resource.Resource {
	return &DomainNameserverResource{}
}

func (r *DomainNameserverResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_nameserver"
}

func (r *DomainNameserverResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom nameserver for a domain in Zone.EU. Before using custom nameservers, ensure the domain's nameservers_custom is set to true via the zoneeu_domain resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The identifier for this resource in format 'domain/hostname'.",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "The domain name this nameserver belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the nameserver (e.g., 'ns1.example.com').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.ListAttribute{
				Description: "List of IP addresses (glue records) for the nameserver. Required when the nameserver hostname is under the same domain.",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *DomainNameserverResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainNameserverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainNameserverResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current nameservers
	currentNS, err := r.client.GetDomainNameservers(data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Domain Nameservers",
			fmt.Sprintf("Could not read nameservers for domain %s: %s", data.Domain.ValueString(), err),
		)
		return
	}

	// Build new nameserver list
	var ips []string
	for _, ip := range data.IP {
		ips = append(ips, ip.ValueString())
	}

	newNS := DomainNameserver{
		Hostname: data.Hostname.ValueString(),
		IP:       ips,
	}

	// Check if this nameserver already exists
	for _, ns := range currentNS {
		if ns.Hostname == data.Hostname.ValueString() {
			resp.Diagnostics.AddError(
				"Nameserver Already Exists",
				fmt.Sprintf("Nameserver %s already exists for domain %s", data.Hostname.ValueString(), data.Domain.ValueString()),
			)
			return
		}
	}

	// Add the new nameserver to the list
	nameservers := append(currentNS, newNS)

	// Create/replace all nameservers
	_, err = r.client.CreateDomainNameservers(data.Domain.ValueString(), nameservers)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Domain Nameserver",
			fmt.Sprintf("Could not create nameserver for domain %s: %s", data.Domain.ValueString(), err),
		)
		return
	}

	// Set the ID
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Domain.ValueString(), data.Hostname.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainNameserverResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ns, err := r.client.GetDomainNameserver(data.Domain.ValueString(), data.Hostname.ValueString())
	if err != nil {
		// If not found, remove from state
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Domain Nameserver",
			fmt.Sprintf("Could not read nameserver %s for domain %s: %s", data.Hostname.ValueString(), data.Domain.ValueString(), err),
		)
		return
	}

	data.Hostname = types.StringValue(ns.Hostname)

	var ips []types.String
	for _, ip := range ns.IP {
		ips = append(ips, types.StringValue(ip))
	}
	data.IP = ips

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Domain.ValueString(), data.Hostname.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainNameserverResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ips []string
	for _, ip := range data.IP {
		ips = append(ips, ip.ValueString())
	}

	ns := &DomainNameserver{
		Hostname: data.Hostname.ValueString(),
		IP:       ips,
	}

	_, err := r.client.UpdateDomainNameserver(data.Domain.ValueString(), data.Hostname.ValueString(), ns)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Domain Nameserver",
			fmt.Sprintf("Could not update nameserver %s for domain %s: %s", data.Hostname.ValueString(), data.Domain.ValueString(), err),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Domain.ValueString(), data.Hostname.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DomainNameserverResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDomainNameserver(data.Domain.ValueString(), data.Hostname.ValueString())
	if err != nil {
		// Ignore 404 errors on delete
		if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") {
			resp.Diagnostics.AddError(
				"Error Deleting Domain Nameserver",
				fmt.Sprintf("Could not delete nameserver %s for domain %s: %s", data.Hostname.ValueString(), data.Domain.ValueString(), err),
			)
			return
		}
	}
}

func (r *DomainNameserverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: domain/hostname
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format 'domain/hostname', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hostname"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
