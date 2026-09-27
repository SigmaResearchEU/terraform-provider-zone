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

var _ resource.Resource = &DNSSSHFPRecordResource{}
var _ resource.ResourceWithImportState = &DNSSSHFPRecordResource{}

func NewDNSSSHFPRecordResource() resource.Resource {
	return &DNSSSHFPRecordResource{}
}

type DNSSSHFPRecordResource struct {
	client *Client
}

type DNSSSHFPRecordResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Zone        types.String `tfsdk:"zone"`
	Name        types.String `tfsdk:"name"`
	Destination types.String `tfsdk:"destination"`
	Algorithm   types.Int64  `tfsdk:"algorithm"`
	Type        types.Int64  `tfsdk:"fingerprint_type"`
	RecordID    types.String `tfsdk:"record_id"`
}

func (r *DNSSSHFPRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_sshfp_record"
}

func (r *DNSSSHFPRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS SSHFP record on Zone.EU.",
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
				Description: "The hostname for the SSHFP record (FQDN, e.g., server.example.com).",
				Required:    true,
			},
			"destination": schema.StringAttribute{
				Description: "The fingerprint in hexadecimal.",
				Required:    true,
			},
			"algorithm": schema.Int64Attribute{
				Description: "The SSH key algorithm: 1=RSA, 2=DSA, 3=ECDSA, 4=Ed25519.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 4),
				},
			},
			"fingerprint_type": schema.Int64Attribute{
				Description: "The fingerprint type: 1=SHA-1, 2=SHA-256.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 2),
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

func (r *DNSSSHFPRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSSSHFPRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSSSHFPRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	record := &DNSRecord{
		Name:        data.Name.ValueString(),
		Destination: data.Destination.ValueString(),
		Algorithm:   int(data.Algorithm.ValueInt64()),
		Type:        int(data.Type.ValueInt64()),
	}

	created, err := r.client.CreateRecord(ctx, recordTypeSSHFP, data.Zone.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create SSHFP record, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Zone.ValueString(), created.ID))
	data.RecordID = types.StringValue(created.ID)

	tflog.Trace(ctx, "created SSHFP record")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSSHFPRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSSSHFPRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, recordID, err := parseRecordID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	record, err := r.client.GetRecord(ctx, recordTypeSSHFP, zone, recordID)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read SSHFP record, got error: %s", err))
		return
	}

	data.Zone = types.StringValue(zone)
	data.Name = types.StringValue(record.Name)
	data.Destination = types.StringValue(record.Destination)
	data.Algorithm = types.Int64Value(int64(record.Algorithm))
	data.Type = types.Int64Value(int64(record.Type))
	data.RecordID = types.StringValue(record.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSSHFPRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSSSHFPRecordResourceModel
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
		Name:        data.Name.ValueString(),
		Destination: data.Destination.ValueString(),
		Algorithm:   int(data.Algorithm.ValueInt64()),
		Type:        int(data.Type.ValueInt64()),
	}

	_, err = r.client.UpdateRecord(ctx, recordTypeSSHFP, zone, recordID, record)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update SSHFP record, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "updated SSHFP record")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSSHFPRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSSSHFPRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, recordID, err := parseRecordID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteRecord(ctx, recordTypeSSHFP, zone, recordID)
	if err != nil {
		// Ignore 404 errors - resource is already deleted
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete SSHFP record, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted SSHFP record")
}

func (r *DNSSSHFPRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
