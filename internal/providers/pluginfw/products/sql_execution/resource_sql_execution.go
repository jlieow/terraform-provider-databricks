package sql_execution

import (
	"context"
	"fmt"
	"reflect"

	"github.com/databricks/databricks-sdk-go/apierr"
	"github.com/databricks/databricks-sdk-go/service/sql"
	"github.com/databricks/terraform-provider-databricks/common"
	pluginfwcommon "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/common"
	pluginfwcontext "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/context"
	"github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/tfschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const resourceName = "sql_execution"

type SqlExecutionModel struct {
	WarehouseId types.String `tfsdk:"warehouse_id"`
	CreateSql   types.String `tfsdk:"create_sql"`
	ReadSql     types.String `tfsdk:"read_sql"`
	DestroySql  types.String `tfsdk:"destroy_sql"`
	ObjectName  types.String `tfsdk:"object_name"`
	ID          types.String `tfsdk:"id"`
	tfschema.Namespace
}

func (m SqlExecutionModel) ApplySchemaCustomizations(s map[string]tfschema.AttributeBuilder) map[string]tfschema.AttributeBuilder {
	s["warehouse_id"] = s["warehouse_id"].SetRequired()
	s["create_sql"] = s["create_sql"].SetRequired()
	s["read_sql"] = s["read_sql"].SetRequired()
	s["destroy_sql"] = s["destroy_sql"].SetRequired()
	s["object_name"] = s["object_name"].SetRequired()
	s["id"] = s["id"].SetComputed().SetOptional()
	s["provider_config"] = s["provider_config"].SetOptional()
	return s
}

func (m SqlExecutionModel) GetComplexFieldTypes(ctx context.Context) map[string]reflect.Type {
	return map[string]reflect.Type{
		"provider_config": reflect.TypeOf(tfschema.ProviderConfig{}),
	}
}

func ResourceSqlExecution() resource.Resource {
	return &sqlExecutionResource{}
}

type sqlExecutionResource struct {
	client *common.DatabricksClient
}

func (r *sqlExecutionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = pluginfwcommon.GetDatabricksProductionName(resourceName)
}

func (r *sqlExecutionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfschema.ResourceStructToSchema(ctx, SqlExecutionModel{}, func(cs tfschema.CustomizableSchema) tfschema.CustomizableSchema {
		cs.AddPlanModifier(stringplanmodifier.RequiresReplace(), "create_sql")
		cs.AddPlanModifier(stringplanmodifier.RequiresReplace(), "object_name")
		cs.AddPlanModifier(stringplanmodifier.UseStateForUnknown(), "id")
		return cs
	})
}

func (r *sqlExecutionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.client == nil && req.ProviderData != nil {
		r.client = pluginfwcommon.ConfigureResource(req, resp)
	}
}

func (r *sqlExecutionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var plan SqlExecutionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID, diags := tfschema.GetWorkspaceIDResource(ctx, plan.ProviderConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, diags := r.client.GetWorkspaceClientForUnifiedProviderWithDiagnostics(ctx, workspaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sqlRes, err := w.StatementExecution.ExecuteStatement(ctx, sql.ExecuteStatementRequest{
		Statement:     plan.CreateSql.ValueString(),
		WarehouseId:   plan.WarehouseId.ValueString(),
		WaitTimeout:   "50s",
		OnWaitTimeout: sql.ExecuteStatementRequestOnWaitTimeoutCancel,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to execute create SQL statement", err.Error())
		return
	}
	if sqlRes.Status.State != sql.StatementStateSucceeded {
		errMsg := string(sqlRes.Status.State)
		if sqlRes.Status.Error != nil {
			errMsg = fmt.Sprintf("%s: %s", sqlRes.Status.State, sqlRes.Status.Error.Message)
		}
		resp.Diagnostics.AddError("create SQL statement did not succeed", errMsg)
		return
	}

	plan.ID = plan.ObjectName
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *sqlExecutionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var state SqlExecutionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID, diags := tfschema.GetWorkspaceIDResource(ctx, state.ProviderConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, diags := r.client.GetWorkspaceClientForUnifiedProviderWithDiagnostics(ctx, workspaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sqlRes, err := w.StatementExecution.ExecuteStatement(ctx, sql.ExecuteStatementRequest{
		Statement:     state.ReadSql.ValueString(),
		WarehouseId:   state.WarehouseId.ValueString(),
		WaitTimeout:   "50s",
		OnWaitTimeout: sql.ExecuteStatementRequestOnWaitTimeoutCancel,
	})
	if err != nil {
		if apierr.IsMissing(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to execute read SQL statement", err.Error())
		return
	}
	if sqlRes.Status.State != sql.StatementStateSucceeded {
		// Statement failed — object does not exist or is not accessible
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *sqlExecutionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	// destroy_sql and warehouse_id are execution-time parameters, not properties
	// of the managed object. Changes to them don't require an API call — just
	// persist the new plan values to state.
	var plan SqlExecutionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *sqlExecutionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var state SqlExecutionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID, diags := tfschema.GetWorkspaceIDResource(ctx, state.ProviderConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, diags := r.client.GetWorkspaceClientForUnifiedProviderWithDiagnostics(ctx, workspaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sqlRes, err := w.StatementExecution.ExecuteStatement(ctx, sql.ExecuteStatementRequest{
		Statement:     state.DestroySql.ValueString(),
		WarehouseId:   state.WarehouseId.ValueString(),
		WaitTimeout:   "50s",
		OnWaitTimeout: sql.ExecuteStatementRequestOnWaitTimeoutCancel,
	})
	if err != nil {
		if apierr.IsMissing(err) {
			return
		}
		resp.Diagnostics.AddError("failed to execute destroy SQL statement", err.Error())
		return
	}
	if sqlRes.Status.State != sql.StatementStateSucceeded {
		errMsg := string(sqlRes.Status.State)
		if sqlRes.Status.Error != nil {
			errMsg = fmt.Sprintf("%s: %s", sqlRes.Status.State, sqlRes.Status.Error.Message)
		}
		resp.Diagnostics.AddError("destroy SQL statement did not succeed", errMsg)
		return
	}
}

func (r *sqlExecutionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("object_name"), req, resp)
}

var _ resource.ResourceWithConfigure = &sqlExecutionResource{}
var _ resource.ResourceWithImportState = &sqlExecutionResource{}
