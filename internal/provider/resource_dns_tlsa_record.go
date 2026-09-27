package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DNSTLSARecordResource{}
var _ resource.ResourceWithImportState = &DNSTLSARecordResource{}

func NewDNSTLSARecordResource() resource.Resource {
	return &DNSTLSARecordResource{}
}

type DNSTLSARecordResource struct {
	client *Client
}

type DNSTLSARecordResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Zone             types.String `tfsdk:"zone"`
	Name             types.String `tfsdk:"name"`
	Destination      types.String `tfsdk:"destination"`
	CertificateUsage types.Int64  `tfsdk:"certificate_usage"`
	Selector         types.Int64  `tfsdk:"selector"`
	MatchingType     types.Int64  `tfsdk:"matching_type"`
	RecordID         types.String `tfsdk:"record_id"`
}

func (r *DNSTLSARecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_tlsa_record"
}

func (r *DNSTLSARecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS TLSA record on Zone.EU.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource in format zone/record_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "The DNS zone name (domain name, e.g., example.com).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The service name (e.g., _443._tcp.example.com for HTTPS).",
				Required:    true,
			},
			"destination": schema.StringAttribute{
				Description: "The certificate association data (hash or full certificate).",
				Required:    true,
			},
			"certificate_usage": schema.Int64Attribute{
				Description: "TLSA certificate usage field (0-3): 0=CA constraint, 1=Service cert constraint, 2=Trust anchor, 3=Domain-issued cert.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 3),
				},
			},
			"selector": schema.Int64Attribute{
				Description: "TLSA selector field (0-1): 0=Full certificate, 1=SubjectPublicKeyInfo.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 1),
				},
			},
			"matching_type": schema.Int64Attribute{
				Description: "TLSA matching type field (0-2): 0=Exact match, 1=SHA-256, 2=SHA-512.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 2),
				},
			},
			"record_id": schema.StringAttribute{
				Description: "The ID of the record in Zone.EU.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DNSTLSARecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DNSTLSARecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSTLSARecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	record := &DNSRecord{
		Name:             data.Name.ValueString(),
		Destination:      data.Destination.ValueString(),
		CertificateUsage: int(data.CertificateUsage.ValueInt64()),
		Selector:         int(data.Selector.ValueInt64()),
		MatchingType:     int(data.MatchingType.ValueInt64()),
	}

	created, err := r.client.CreateRecord(ctx, recordTypeTLSA, data.Zone.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create TLSA record, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Zone.ValueString(), created.ID))
	data.RecordID = types.StringValue(created.ID)

	tflog.Trace(ctx, "created TLSA record")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSTLSARecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSTLSARecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, recordID, err := parseRecordID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	record, err := r.client.GetRecord(ctx, recordTypeTLSA, zone, recordID)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TLSA record, got error: %s", err))
		return
	}

	data.Zone = types.StringValue(zone)
	data.Name = types.StringValue(record.Name)
	data.Destination = types.StringValue(record.Destination)
	data.CertificateUsage = types.Int64Value(int64(record.CertificateUsage))
	data.Selector = types.Int64Value(int64(record.Selector))
	data.MatchingType = types.Int64Value(int64(record.MatchingType))
	data.RecordID = types.StringValue(record.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSTLSARecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSTLSARecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, recordID, err := parseRecordID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	record := &DNSRecord{
		Name:             data.Name.ValueString(),
		Destination:      data.Destination.ValueString(),
		CertificateUsage: int(data.CertificateUsage.ValueInt64()),
		Selector:         int(data.Selector.ValueInt64()),
		MatchingType:     int(data.MatchingType.ValueInt64()),
	}

	_, err = r.client.UpdateRecord(ctx, recordTypeTLSA, zone, recordID, record)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update TLSA record, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "updated TLSA record")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSTLSARecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSTLSARecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, recordID, err := parseRecordID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteRecord(ctx, recordTypeTLSA, zone, recordID)
	if err != nil {
		// Ignore 404 errors - resource is already deleted
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete TLSA record, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted TLSA record")
}

func (r *DNSTLSARecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: zone/record_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("record_id"), parts[1])...)
}
