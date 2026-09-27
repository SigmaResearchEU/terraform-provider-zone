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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DNSNSRecordResource{}
var _ resource.ResourceWithImportState = &DNSNSRecordResource{}

func NewDNSNSRecordResource() resource.Resource {
	return &DNSNSRecordResource{}
}

type DNSNSRecordResource struct {
	client *Client
}

type DNSNSRecordResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Zone          types.String `tfsdk:"zone"`
	Name          types.String `tfsdk:"name"`
	Destination   types.String `tfsdk:"destination"`
	RecordID      types.String `tfsdk:"record_id"`
	ForceRecreate types.Bool   `tfsdk:"force_recreate"`
}

func (r *DNSNSRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_ns_record"
}

func (r *DNSNSRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS NS record on Zone.EU.",
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
				Description: "The hostname for the NS record (FQDN, e.g., subdomain.example.com).",
				Required:    true,
			},
			"destination": schema.StringAttribute{
				Description: "The nameserver hostname.",
				Required:    true,
			},
			"record_id": schema.StringAttribute{
				Description: "The ID of the record in Zone.EU.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"force_recreate": schema.BoolAttribute{
				Description: "If true, delete existing record with same name before creating. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *DNSNSRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSNSRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If force_recreate is true, check for existing record and update it instead of creating
	if data.ForceRecreate.ValueBool() {
		existing, err := r.client.FindNSRecordByNameWithContext(ctx, data.Zone.ValueString(), data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to check for existing NS record, got error: %s", err))
			return
		}
		if existing != nil {
			tflog.Info(ctx, "force_recreate: updating existing NS record instead of creating new", map[string]interface{}{
				"zone":      data.Zone.ValueString(),
				"name":      data.Name.ValueString(),
				"record_id": existing.ID,
			})

			record := &DNSRecord{
				Name:        data.Name.ValueString(),
				Destination: data.Destination.ValueString(),
			}

			updated, err := r.client.UpdateNSRecordWithContext(ctx, data.Zone.ValueString(), existing.ID, record)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update existing NS record for force_recreate, got error: %s", err))
				return
			}

			data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Zone.ValueString(), updated.ID))
			data.RecordID = types.StringValue(updated.ID)

			tflog.Trace(ctx, "updated existing NS record via force_recreate")
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	record := &DNSRecord{
		Name:        data.Name.ValueString(),
		Destination: data.Destination.ValueString(),
	}

	created, err := r.client.CreateNSRecordWithContext(ctx, data.Zone.ValueString(), record)
	if err != nil {
		// Handle zone_conflict by adopting existing record into state
		if strings.Contains(err.Error(), "zone_conflict") {
			tflog.Info(ctx, "Record already exists (zone_conflict), adopting into state", map[string]interface{}{
				"zone": data.Zone.ValueString(),
				"name": data.Name.ValueString(),
			})
			existing, findErr := r.client.FindNSRecordByNameWithContext(ctx, data.Zone.ValueString(), data.Name.ValueString())
			if findErr == nil && existing != nil {
				data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Zone.ValueString(), existing.ID))
				data.RecordID = types.StringValue(existing.ID)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			// If we couldn't find/adopt, fall through to error
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create NS record, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Zone.ValueString(), created.ID))
	data.RecordID = types.StringValue(created.ID)

	tflog.Trace(ctx, "created NS record")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, recordID, err := parseRecordID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	record, err := r.client.GetNSRecordWithContext(ctx, zone, recordID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read NS record, got error: %s", err))
		return
	}

	data.Zone = types.StringValue(zone)
	data.Name = types.StringValue(record.Name)
	data.Destination = types.StringValue(record.Destination)
	data.RecordID = types.StringValue(record.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSNSRecordResourceModel
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
	}

	_, err = r.client.UpdateNSRecordWithContext(ctx, zone, recordID, record)
	if err != nil {
		// Handle zone_conflict when force_recreate is enabled
		if strings.Contains(err.Error(), "zone_conflict") && data.ForceRecreate.ValueBool() {
			tflog.Info(ctx, "zone_conflict during update with force_recreate=true, deleting all duplicates and recreating")
			
			// Find and delete ALL records with this name (handles duplicates)
			allRecords, findErr := r.client.FindAllNSRecordsByNameWithContext(ctx, zone, data.Name.ValueString())
			if findErr != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find existing records: %s", findErr))
				return
			}
			
			// Delete all matching records
			for _, rec := range allRecords {
				deleteErr := r.client.DeleteNSRecordWithContext(ctx, zone, rec.ID)
				if deleteErr != nil {
					// Ignore 404 errors
					if !strings.Contains(deleteErr.Error(), "404") {
						tflog.Warn(ctx, fmt.Sprintf("Failed to delete duplicate record %s: %s", rec.ID, deleteErr))
					}
				}
			}
			
			// Create fresh record
			created, createErr := r.client.CreateNSRecordWithContext(ctx, zone, record)
			if createErr != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to recreate NS record after deleting duplicates: %s", createErr))
				return
			}
			
			// Update state with new record ID
			data.RecordID = types.StringValue(created.ID)
			data.ID = types.StringValue(fmt.Sprintf("%s/%s", zone, created.ID))
			tflog.Trace(ctx, "recreated NS record after deleting duplicates")
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
		
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update NS record, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "updated NS record")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, recordID, err := parseRecordID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteNSRecordWithContext(ctx, zone, recordID)
	if err != nil {
		// Ignore 404 errors - resource is already deleted
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete NS record, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted NS record")
}

func (r *DNSNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
