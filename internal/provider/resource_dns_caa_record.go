package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DNSCAARecordResource{}
var _ resource.ResourceWithImportState = &DNSCAARecordResource{}

func NewDNSCAARecordResource() resource.Resource {
	return &DNSCAARecordResource{}
}

type DNSCAARecordResource struct {
	client *Client
}

type DNSCAARecordResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Zone        types.String `tfsdk:"zone"`
	Name        types.String `tfsdk:"name"`
	Destination types.String `tfsdk:"destination"`
	Flag        types.Int64  `tfsdk:"flag"`
	Tag         types.String `tfsdk:"tag"`
	RecordID    types.String `tfsdk:"record_id"`
}

func (r *DNSCAARecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_caa_record"
}

func (r *DNSCAARecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS CAA record on Zone.EU.",
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
				Description: "The hostname for the CAA record (FQDN, e.g., example.com).",
				Required:    true,
			},
			"destination": schema.StringAttribute{
				Description: "The value associated with the tag (e.g., CA domain).",
				Required:    true,
			},
			"flag": schema.Int64Attribute{
				Description: "The CAA record flag (0-255). Commonly 0 for non-critical or 128 for critical.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 255),
				},
			},
			"tag": schema.StringAttribute{
				Description: "The CAA tag: issue, issuewild, or iodef.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("issue", "issuewild", "iodef"),
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

func (r *DNSCAARecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSCAARecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSCAARecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	record := &DNSRecord{
		Name:        data.Name.ValueString(),
		Destination: data.Destination.ValueString(),
		Flag:        int(data.Flag.ValueInt64()),
		Tag:         data.Tag.ValueString(),
	}

	created, err := r.client.CreateRecord(ctx, recordTypeCAA, data.Zone.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create CAA record, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Zone.ValueString(), created.ID))
	data.RecordID = types.StringValue(created.ID)

	tflog.Trace(ctx, "created CAA record")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSCAARecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSCAARecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, recordID, err := parseRecordID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	record, err := r.client.GetRecord(ctx, recordTypeCAA, zone, recordID)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read CAA record, got error: %s", err))
		return
	}

	data.Zone = types.StringValue(zone)
	data.Name = types.StringValue(record.Name)
	data.Destination = types.StringValue(record.Destination)
	data.Flag = types.Int64Value(int64(record.Flag))
	data.Tag = types.StringValue(record.Tag)
	data.RecordID = types.StringValue(record.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSCAARecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSCAARecordResourceModel
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
		Flag:        int(data.Flag.ValueInt64()),
		Tag:         data.Tag.ValueString(),
	}

	_, err = r.client.UpdateRecord(ctx, recordTypeCAA, zone, recordID, record)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update CAA record, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "updated CAA record")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSCAARecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSCAARecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, recordID, err := parseRecordID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteRecord(ctx, recordTypeCAA, zone, recordID)
	if err != nil {
		// Ignore 404 errors - resource is already deleted
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete CAA record, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted CAA record")
}

func (r *DNSCAARecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
