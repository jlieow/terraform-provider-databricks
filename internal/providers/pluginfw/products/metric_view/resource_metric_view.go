package metric_view

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/databricks/databricks-sdk-go/apierr"
	"github.com/databricks/databricks-sdk-go/service/sql"
	"github.com/databricks/terraform-provider-databricks/common"
	pluginfwcommon "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/common"
	pluginfwcontext "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/context"
	"github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/tfschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const resourceName = "metric_view"
const sqlExecWaitTimeout = "50s"

var _ resource.ResourceWithConfigure = &metricViewResource{}
var _ resource.ResourceWithImportState = &metricViewResource{}

type MetricViewModel struct {
	WarehouseId       types.String `tfsdk:"warehouse_id"`
	Name              types.String `tfsdk:"name"`
	CatalogName       types.String `tfsdk:"catalog_name"`
	SchemaName        types.String `tfsdk:"schema_name"`
	YamlSpecification types.String `tfsdk:"yaml_specification"`
	ID                types.String `tfsdk:"id"`
	tfschema.Namespace
}

func (m MetricViewModel) ApplySchemaCustomizations(s map[string]tfschema.AttributeBuilder) map[string]tfschema.AttributeBuilder {
	s["warehouse_id"] = s["warehouse_id"].SetRequired()
	s["name"] = s["name"].SetRequired()
	s["catalog_name"] = s["catalog_name"].SetRequired()
	s["schema_name"] = s["schema_name"].SetRequired()
	s["yaml_specification"] = s["yaml_specification"].SetRequired()
	s["id"] = s["id"].SetComputed().SetOptional()
	s["provider_config"] = s["provider_config"].SetOptional()
	return s
}

func (m MetricViewModel) GetComplexFieldTypes(ctx context.Context) map[string]reflect.Type {
	return map[string]reflect.Type{
		"provider_config": reflect.TypeOf(tfschema.ProviderConfig{}),
	}
}

func (m MetricViewModel) fullName() string {
	return fmt.Sprintf("%s.%s.%s", m.CatalogName.ValueString(), m.SchemaName.ValueString(), m.Name.ValueString())
}

func (m MetricViewModel) sqlFullName() string {
	return fmt.Sprintf("`%s`.`%s`.`%s`", m.CatalogName.ValueString(), m.SchemaName.ValueString(), m.Name.ValueString())
}

func ResourceMetricView() resource.Resource {
	return &metricViewResource{}
}

type metricViewResource struct {
	client *common.DatabricksClient
}

func (r *metricViewResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = pluginfwcommon.GetDatabricksProductionName(resourceName)
}

func (r *metricViewResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfschema.ResourceStructToSchema(ctx, MetricViewModel{}, func(cs tfschema.CustomizableSchema) tfschema.CustomizableSchema {
		cs.AddPlanModifier(stringplanmodifier.RequiresReplace(), "name")
		cs.AddPlanModifier(stringplanmodifier.RequiresReplace(), "catalog_name")
		cs.AddPlanModifier(stringplanmodifier.RequiresReplace(), "schema_name")
		cs.AddPlanModifier(stringplanmodifier.UseStateForUnknown(), "id")
		cs.AddPlanModifier(suppressCaseChangePlanModifier{}, "name")
		cs.AddPlanModifier(suppressCaseChangePlanModifier{}, "catalog_name")
		cs.AddPlanModifier(suppressCaseChangePlanModifier{}, "schema_name")
		return cs
	})
}

func (r *metricViewResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.client == nil && req.ProviderData != nil {
		r.client = pluginfwcommon.ConfigureResource(req, resp)
	}
}

func (r *metricViewResource) executeSql(ctx context.Context, warehouseID string, providerConfig types.Object, statement string) error {
	workspaceID, diags := tfschema.GetWorkspaceIDResource(ctx, providerConfig)
	if diags.HasError() {
		return fmt.Errorf("failed to get workspace ID: %s", diags.Errors()[0].Detail())
	}

	w, diags := r.client.GetWorkspaceClientForUnifiedProviderWithDiagnostics(ctx, workspaceID)
	if diags.HasError() {
		return fmt.Errorf("failed to get workspace client: %s", diags.Errors()[0].Detail())
	}

	sqlRes, err := w.StatementExecution.ExecuteStatement(ctx, sql.ExecuteStatementRequest{
		Statement:     statement,
		WarehouseId:   warehouseID,
		WaitTimeout:   sqlExecWaitTimeout,
		OnWaitTimeout: sql.ExecuteStatementRequestOnWaitTimeoutCancel,
	})
	if err != nil {
		return err
	}
	if sqlRes.Status.State != sql.StatementStateSucceeded {
		errMsg := string(sqlRes.Status.State)
		if sqlRes.Status.Error != nil {
			errMsg = fmt.Sprintf("%s: %s", sqlRes.Status.State, sqlRes.Status.Error.Message)
		}
		return fmt.Errorf("statement failed to execute: %s", errMsg)
	}
	return nil
}

func (r *metricViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var plan MetricViewModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stmt := fmt.Sprintf("CREATE METRIC VIEW %s AS\n%s;", plan.sqlFullName(), plan.YamlSpecification.ValueString())
	err := r.executeSql(ctx, plan.WarehouseId.ValueString(), plan.ProviderConfig, stmt)
	if err != nil {
		resp.Diagnostics.AddError("failed to create metric view", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.fullName())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *metricViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var state MetricViewModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stmt := fmt.Sprintf("DESCRIBE EXTENDED %s;", state.sqlFullName())
	err := r.executeSql(ctx, state.WarehouseId.ValueString(), state.ProviderConfig, stmt)
	if err != nil {
		if apierr.IsMissing(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		// If the SQL statement fails, the metric view likely doesn't exist
		if strings.Contains(err.Error(), "TABLE_OR_VIEW_NOT_FOUND") || strings.Contains(err.Error(), "RESOURCE_DOES_NOT_EXIST") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to read metric view", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *metricViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var plan MetricViewModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stmt := fmt.Sprintf("CREATE OR REPLACE METRIC VIEW %s AS\n%s;", plan.sqlFullName(), plan.YamlSpecification.ValueString())
	err := r.executeSql(ctx, plan.WarehouseId.ValueString(), plan.ProviderConfig, stmt)
	if err != nil {
		resp.Diagnostics.AddError("failed to update metric view", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *metricViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var state MetricViewModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stmt := fmt.Sprintf("DROP METRIC VIEW IF EXISTS %s;", state.sqlFullName())
	err := r.executeSql(ctx, state.WarehouseId.ValueString(), state.ProviderConfig, stmt)
	if err != nil {
		if apierr.IsMissing(err) {
			return
		}
		resp.Diagnostics.AddError("failed to delete metric view", err.Error())
		return
	}
}

func (r *metricViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID is the full name: catalog.schema.name
	parts := strings.SplitN(req.ID, ".", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"invalid import ID",
			fmt.Sprintf("expected format: catalog_name.schema_name.name, got: %s", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("catalog_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schema_name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// suppressCaseChangePlanModifier suppresses diffs when only casing changes.
type suppressCaseChangePlanModifier struct{}

func (m suppressCaseChangePlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs when only character casing changes."
}

func (m suppressCaseChangePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m suppressCaseChangePlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}
	if strings.EqualFold(req.StateValue.ValueString(), req.PlanValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
