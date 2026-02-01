package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/databricks/databricks-sdk-go/apierr"
	"github.com/databricks/databricks-sdk-go/service/apps"
	"github.com/databricks/terraform-provider-databricks/common"
	pluginfwcommon "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/common"
	pluginfwcontext "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const appDeploymentResourceName = "app_deployment"

type AppDeploymentModel struct {
	AppName        types.String `tfsdk:"app_name"`
	SourceCodePath types.String `tfsdk:"source_code_path"`
	Triggers       types.Map    `tfsdk:"triggers"`
	DeploymentId   types.String `tfsdk:"deployment_id"`
}

func ResourceAppDeployment() resource.Resource {
	return &resourceAppDeployment{}
}

type resourceAppDeployment struct {
	client *common.DatabricksClient
}

func (r *resourceAppDeployment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = pluginfwcommon.GetDatabricksProductionName(appDeploymentResourceName)
}

func (r *resourceAppDeployment) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a deployment for a Databricks app. All configuration changes force replacement.",
		Attributes: map[string]schema.Attribute{
			"app_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the app to deploy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_code_path": schema.StringAttribute{
				Required:    true,
				Description: "The workspace file system path of the source code.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"triggers": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Arbitrary map of string values that, when changed, will trigger a new deployment.",
				PlanModifiers: []planmodifier.Map{
					mapRequiresReplace{},
				},
			},
			"deployment_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the deployment.",
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
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, appDeploymentResourceName)

	var plan AppDeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.client.WorkspaceClient()
	if err != nil {
		resp.Diagnostics.AddError("failed to get workspace client", err.Error())
		return
	}

	waiter, err := w.Apps.Deploy(ctx, apps.CreateAppDeploymentRequest{
		AppName: plan.AppName.ValueString(),
		AppDeployment: apps.AppDeployment{
			SourceCodePath: plan.SourceCodePath.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to deploy app", err.Error())
		return
	}

	deployment, err := waiter.Get()
	if err != nil {
		resp.Diagnostics.AddError("failed to wait for deployment to succeed", err.Error())
		return
	}

	plan.DeploymentId = types.StringValue(deployment.DeploymentId)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAppDeployment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, appDeploymentResourceName)

	var state AppDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName, deploymentId, err := parseAppDeploymentId(state)
	if err != nil {
		resp.Diagnostics.AddError("invalid resource ID", err.Error())
		return
	}

	w, err := r.client.WorkspaceClient()
	if err != nil {
		resp.Diagnostics.AddError("failed to get workspace client", err.Error())
		return
	}

	deployment, err := w.Apps.GetDeploymentByAppNameAndDeploymentId(ctx, appName, deploymentId)
	if err != nil {
		if apierr.IsMissing(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to read deployment", err.Error())
		return
	}

	state.DeploymentId = types.StringValue(deployment.DeploymentId)
	state.SourceCodePath = types.StringValue(deployment.SourceCodePath)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceAppDeployment) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields have RequiresReplace, so Update should never be called.
	resp.Diagnostics.AddError("unexpected update", "all fields force replacement; update should not be called")
}

func (r *resourceAppDeployment) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Deployments cannot be deleted, only superseded. No-op.
}

func parseAppDeploymentId(state AppDeploymentModel) (string, string, error) {
	appName := state.AppName.ValueString()
	deploymentId := state.DeploymentId.ValueString()
	if appName == "" || deploymentId == "" {
		return "", "", fmt.Errorf("both app_name and deployment_id must be set, got app_name=%q deployment_id=%q", appName, deploymentId)
	}
	return appName, deploymentId, nil
}

func (r *resourceAppDeployment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "|", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("invalid import ID", fmt.Sprintf("expected format: app_name|deployment_id, got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, AppDeploymentModel{
		AppName:      types.StringValue(parts[0]),
		DeploymentId: types.StringValue(parts[1]),
	})...)
}

// mapRequiresReplace is a plan modifier that forces replacement when the map value changes.
type mapRequiresReplace struct{}

func (m mapRequiresReplace) Description(_ context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (m mapRequiresReplace) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m mapRequiresReplace) PlanModifyMap(_ context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}
	if !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}

var _ resource.ResourceWithConfigure = &resourceAppDeployment{}
var _ resource.ResourceWithImportState = &resourceAppDeployment{}
