package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.com/rackn/provision/v4/api"
	"gitlab.com/rackn/provision/v4/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MachineResource{}
var _ resource.ResourceWithImportState = &MachineResource{}

func NewMachineResource() resource.Resource {
	return &MachineResource{}
}

// MachineResource defines the resource implementation.
type MachineResource struct {
	session *api.Client
}

// MachineResourceModel describes the resource data model.
type MachineResourceModel struct {
	Id types.String `tfsdk:"id"`

	Pool               types.String `tfsdk:"pool"`
	AllocateWorkflow   types.String `tfsdk:"allocate_workflow"`
	DeallocateWorkflow types.String `tfsdk:"deallocate_workflow"`
	Timeout            types.String `tfsdk:"timeout"`
	AddProfiles        types.List   `tfsdk:"add_profiles"`
	AddParameters      types.List   `tfsdk:"add_parameters"`
	Filters            types.List   `tfsdk:"filters"`
	AuthorizedKeys     types.List   `tfsdk:"authorized_keys"`

	Address types.String `tfsdk:"address"`
	Name    types.String `tfsdk:"name"`
	Status  types.String `tfsdk:"status"`
}

func (r *MachineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

func (r *MachineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Machine resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"pool": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Pool to operate against for machine actions",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allocate_workflow": schema.StringAttribute{
				MarkdownDescription: "Workflow to run when the machine is allocated in the pool",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deallocate_workflow": schema.StringAttribute{
				MarkdownDescription: "Workflow to run when the machine is released to the pool",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeout": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Maximum time to wait for the machine to complete transition.  Time string format.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"add_profiles": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of profiles to add to the machine when allocating.  Profiles are removed on release.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"add_parameters": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of parameters to add to the machine when allocating.  Parameters are removed on release.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"filters": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of filters to restrict the search for a machie (usee Digital Rebar format e.g. FilterVar=Fn(value))",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"authorized_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of ssh public keys that should be added to the access-keys parameter on the machine.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Returns the IP address on the machine, Machine.Address field",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Returns the Pool status of the machine, Machine.PoolStatus field",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Returns the Name of the machine, Machine.Name field",
			},
		},
	}
}

func (r *MachineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Config)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.session = client.session
}

func (r *MachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "[resourceMachineAllocate] Allocating new drp_machine")
	var plan MachineResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool := "default"
	if p := plan.Pool.ValueString(); p != "" {
		pool = p
	}
	plan.Pool = types.StringValue(pool)

	timeout := "5m"
	if t := plan.Timeout.ValueString(); t != "" {
		timeout = t
	}
	parms := map[string]interface{}{
		"pool/wait-timeout": timeout,
	}
	plan.Timeout = types.StringValue(timeout)

	pwf := plan.AllocateWorkflow.ValueString()
	if pwf != "" {
		parms["pool/workflow"] = pwf
	}

	profiles := []string{}
	diags = plan.AddProfiles.ElementsAs(ctx, profiles, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(profiles) > 0 {
		parms["pool/add-profiles"] = profiles
	}

	parameters := map[string]interface{}{}
	akeys := []string{}
	diags = plan.AuthorizedKeys.ElementsAs(ctx, akeys, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(akeys) > 0 {
		accesskeys := map[string]string{}
		for i, p := range akeys {
			accesskeys[fmt.Sprintf("terraform-%d", i)] = p
		}
		parameters["access-keys"] = accesskeys
	}

	aparams := []string{}
	diags = plan.AddParameters.ElementsAs(ctx, aparams, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, p := range aparams {
		param := strings.Split(p, ":")
		if len(param) < 2 {
			resp.Diagnostics.AddError("add_parameter format not correct", p)
			return
		}
		key := param[0]
		value := strings.TrimLeft(param[1], " ")
		parameters[key] = value
	}
	if len(parameters) > 0 {
		parms["pool/add-parameters"] = parameters
	}
	allFilters := []string{"Runnable=Eq(true)", "WorkflowComplete=Eq(true)", "WorkOrderMode=Eq(false)"}
	filters := []string{}
	diags = plan.Filters.ElementsAs(ctx, filters, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, f := range filters {
		allFilters = append(allFilters, f)
	}
	parms["pool/filter"] = allFilters

	pr := []*models.PoolResult{}
	creq := r.session.Req().Post(parms).UrlFor("pools", pool, "allocateMachines")
	if err := creq.Do(&pr); err != nil {
		tflog.Debug(ctx, fmt.Sprintf("POST error %+v | %+v", err, creq))
		resp.Diagnostics.AddError(fmt.Sprintf("Error allocated from pool %s: %s", pool, err), "")
		return
	}
	mc := pr[0]
	tflog.Debug(ctx, fmt.Sprintf("Allocated %s machine %s (%s)", mc.Status, mc.Name, mc.Uuid))
	plan.Status = types.StringValue(string(mc.Status))
	plan.Name = types.StringValue(mc.Name)
	plan.Id = types.StringValue(mc.Uuid)

	if mo, err := r.session.GetModel("machines", mc.Uuid); err == nil {
		machineObject := mo.(*models.Machine)
		plan.Address = types.StringValue(machineObject.Address.String())
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to lookup machine: %v", err))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "[resourceMachineRead] Reading drp_machine")

	var plan MachineResourceModel

	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := plan.Id.ValueString()
	if uuid == "" {
		tflog.Debug(ctx, "Requires Uuuid from id")
		resp.Diagnostics.AddError("Requires Uuid from id", "")
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading machine %s", uuid))
	mo, err := r.session.GetModel("machines", uuid)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("[resourceMachineRead] Unable to get machine: %s", uuid))
		resp.Diagnostics.AddError(fmt.Sprintf("Unable to get machine: %s", uuid), "")
		return
	}
	machineObject := mo.(*models.Machine)
	if machineObject.PoolStatus == "HoldBuild" {
		tflog.Debug(ctx, "Machine: #{uuid} in HoldBuild status. Investigate the cause on the DRP endpoint.")
		resp.Diagnostics.AddError(fmt.Sprintf("machine %s stuck in HoldBuild status", uuid), "")
		return
	}

	plan.Status = types.StringValue(string(machineObject.PoolStatus))
	plan.Name = types.StringValue(machineObject.Name)
	plan.Id = types.StringValue(machineObject.Uuid.String())
	plan.Address = types.StringValue(machineObject.Address.String())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "[resourceMachineUpdate] Updating drp_machine")

	var plan MachineResourceModel

	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := plan.Id.ValueString()
	if uuid == "" {
		tflog.Debug(ctx, "Requires Uuuid from id")
		resp.Diagnostics.AddError("Requires Uuid from id", "")
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading machine %s", uuid))
	mo, err := r.session.GetModel("machines", uuid)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("[resourceMachineRead] Unable to get machine: %s", uuid))
		resp.Diagnostics.AddError(fmt.Sprintf("Unable to get machine: %s", uuid), "")
		return
	}
	machineObject := mo.(*models.Machine)
	if machineObject.PoolStatus == "HoldBuild" {
		tflog.Debug(ctx, "Machine: #{uuid} in HoldBuild status. Investigate the cause on the DRP endpoint.")
		resp.Diagnostics.AddError(fmt.Sprintf("machine %s stuck in HoldBuild status", uuid), "")
		return
	}

	plan.Status = types.StringValue(string(machineObject.PoolStatus))
	plan.Name = types.StringValue(machineObject.Name)
	plan.Id = types.StringValue(machineObject.Uuid.String())
	plan.Address = types.StringValue(machineObject.Address.String())
}

func (r *MachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "[resourceMachineAllocate] Releasing drp_machine")
	var plan MachineResourceModel

	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := plan.Id.ValueString()
	if uuid == "" {
		tflog.Debug(ctx, "Requires Uuuid from id")
		resp.Diagnostics.AddError("Requires Uuid from id", "")
		return
	}

	pool := "default"
	if p := plan.Pool.ValueString(); p != "" {
		pool = p
	}
	plan.Pool = types.StringValue(pool)

	timeout := "5m"
	if t := plan.Timeout.ValueString(); t != "" {
		timeout = t
	}
	parms := map[string]interface{}{
		"pool/wait-timeout": timeout,
	}
	plan.Timeout = types.StringValue(timeout)

	pwf := plan.DeallocateWorkflow.ValueString()
	if pwf != "" {
		parms["pool/workflow"] = pwf
	}

	profiles := []string{}
	diags = plan.AddProfiles.ElementsAs(ctx, profiles, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(profiles) > 0 {
		parms["pool/remove-profiles"] = profiles
	}

	parameters := []string{}
	akeys := []string{}
	diags = plan.AuthorizedKeys.ElementsAs(ctx, akeys, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(akeys) > 0 {
		parameters = append(parameters, "access-keys")
	}

	aparams := []string{}
	diags = plan.AddParameters.ElementsAs(ctx, aparams, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, p := range aparams {
		param := strings.Split(p, ":")
		if len(param) < 2 {
			resp.Diagnostics.AddError("add_parameter format not correct", p)
			return
		}
		key := param[0]
		parameters = append(parameters, key)
	}
	if len(parameters) > 0 {
		parms["pool/remove-parameters"] = parameters
	}

	pr := []*models.PoolResult{}
	creq := r.session.Req().Post(parms).UrlFor("pools", pool, "releaseMachines")
	if err := creq.Do(&pr); err != nil {
		tflog.Error(ctx, fmt.Sprintf("[resourceMachineDelete] POST error %+v | %+v", err, creq))
		resp.Diagnostics.AddError(fmt.Sprintf("Error releasing %s from pool %s: %s", uuid, pool, err), "")
		return
	}

	mc := pr[0]
	if mc.Status == "Free" {
		plan.Status = types.StringValue(string(mc.Status))
		plan.Name = types.StringValue(uuid)
		plan.Id = types.StringValue("")
		plan.Address = types.StringValue("")

		diags = resp.State.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)
	} else {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not release %s from pool %s", uuid, pool), "")
		return
	}
}

func (r *MachineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
