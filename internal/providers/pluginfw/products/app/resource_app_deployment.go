package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/databricks/databricks-sdk-go/service/apps"
	"github.com/databricks/terraform-provider-databricks/common"
	pluginfwcommon "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/common"
	pluginfwcontext "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const deploymentResourceName = "app_deployment"

func ResourceAppDeployment() resource.Resource {
	return &resourceAppDeployment{}
}

type appDeploymentModel struct {
	AppName        types.String `tfsdk:"app_name"`
	SourceCodePath types.String `tfsdk:"source_code_path"`
	Mode           types.String `tfsdk:"mode"`
	Triggers       types.Map    `tfsdk:"triggers"`
	DeploymentId   types.String `tfsdk:"deployment_id"`
	Status         types.String `tfsdk:"status"`
	CreateTime     types.String `tfsdk:"create_time"`
}

type resourceAppDeployment struct {
	client *common.DatabricksClient
}

func (r resourceAppDeployment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = pluginfwcommon.GetDatabricksProductionName(deploymentResourceName)
}

func (r resourceAppDeployment) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deploys code to a Databricks app.",
		Attributes: map[string]schema.Attribute{
			"app_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the app to deploy to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_code_path": schema.StringAttribute{
				Required:    true,
				Description: "The workspace filesystem path of the source code to deploy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Optional:    true,
				Description: "The deployment mode. Allowed values are SNAPSHOT and AUTO_SYNC.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"triggers": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "A map of arbitrary strings that, when changed, will force a new deployment.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"deployment_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique ID of the deployment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the deployment.",
			},
			"create_time": schema.StringAttribute{
				Computed:    true,
				Description: "The creation time of the deployment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceAppDeployment) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.client == nil && req.ProviderData != nil {
		r.client = pluginfwcommon.ConfigureResource(req, resp)
	}
}

func (r *resourceAppDeployment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, deploymentResourceName)

	var plan appDeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.client.WorkspaceClient()
	if err != nil {
		resp.Diagnostics.AddError("failed to get workspace client", err.Error())
		return
	}

	deployReq := apps.CreateAppDeploymentRequest{
		AppName: plan.AppName.ValueString(),
		AppDeployment: apps.AppDeployment{
			SourceCodePath: plan.SourceCodePath.ValueString(),
		},
	}
	if !plan.Mode.IsNull() && !plan.Mode.IsUnknown() {
		deployReq.AppDeployment.Mode = apps.AppDeploymentMode(plan.Mode.ValueString())
	}

	deployment, err := w.Apps.DeployAndWait(ctx, deployReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to deploy app", err.Error())
		return
	}

	plan.DeploymentId = types.StringValue(deployment.DeploymentId)
	plan.CreateTime = types.StringValue(deployment.CreateTime)
	if deployment.Status != nil {
		plan.Status = types.StringValue(string(deployment.Status.State))
	} else {
		plan.Status = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAppDeployment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, deploymentResourceName)

	var state appDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.client.WorkspaceClient()
	if err != nil {
		resp.Diagnostics.AddError("failed to get workspace client", err.Error())
		return
	}

	deployment, err := w.Apps.GetDeploymentByAppNameAndDeploymentId(ctx, state.AppName.ValueString(), state.DeploymentId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read app deployment", err.Error())
		return
	}

	state.SourceCodePath = types.StringValue(deployment.SourceCodePath)
	state.CreateTime = types.StringValue(deployment.CreateTime)
	if deployment.Mode != "" {
		state.Mode = types.StringValue(string(deployment.Mode))
	}
	if deployment.Status != nil {
		state.Status = types.StringValue(string(deployment.Status.State))
	} else {
		state.Status = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceAppDeployment) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All user-settable fields are ForceNew, so Update should never be called.
	resp.Diagnostics.AddError("unexpected update", "app_deployment does not support in-place updates")
}

func (r *resourceAppDeployment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Deployments cannot be deleted via the API. Removing from state is sufficient.
}

func (r *resourceAppDeployment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("invalid import ID", fmt.Sprintf("expected format: app_name/deployment_id, got: %s", req.ID))
		return
	}

	var state appDeploymentModel
	state.AppName = types.StringValue(parts[0])
	state.DeploymentId = types.StringValue(parts[1])
	state.Triggers = types.MapNull(types.StringType)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

var _ resource.ResourceWithConfigure = &resourceAppDeployment{}
var _ resource.ResourceWithImportState = &resourceAppDeployment{}
