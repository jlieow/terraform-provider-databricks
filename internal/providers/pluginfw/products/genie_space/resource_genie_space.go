package genie_space

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/databricks/databricks-sdk-go/apierr"
	"github.com/databricks/databricks-sdk-go/service/dashboards"
	"github.com/databricks/terraform-provider-databricks/common"
	pluginfwcommon "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/common"
	pluginfwcontext "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/context"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const resourceName = "genie_space"

var _ resource.ResourceWithConfigure = &GenieSpaceResource{}
var _ resource.ResourceWithImportState = &GenieSpaceResource{}

func ResourceGenieSpace() resource.Resource {
	return &GenieSpaceResource{}
}

type GenieSpaceResource struct {
	Client *common.DatabricksClient
}

type GenieSpaceModel struct {
	ID              types.String `tfsdk:"id"`
	Title           types.String `tfsdk:"title"`
	WarehouseId     types.String `tfsdk:"warehouse_id"`
	Description     types.String `tfsdk:"description"`
	ParentPath      types.String `tfsdk:"parent_path"`
	SpaceId         types.String `tfsdk:"space_id"`
	Tables          types.List   `tfsdk:"tables"`
	SampleQuestions types.List   `tfsdk:"sample_questions"`
}

type serializedSpaceContent struct {
	Version     int                        `json:"version"`
	Config      serializedSpaceConfig      `json:"config"`
	DataSources serializedSpaceDataSources `json:"data_sources"`
}

type serializedSpaceConfig struct {
	SampleQuestions []serializedSampleQuestion `json:"sample_questions,omitempty"`
}

type serializedSpaceDataSources struct {
	Tables []serializedTable `json:"tables,omitempty"`
}

type serializedTable struct {
	Identifier string `json:"identifier"`
}

type serializedSampleQuestion struct {
	ID       string   `json:"id"`
	Question []string `json:"question"`
}

func (r *GenieSpaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = pluginfwcommon.GetDatabricksProductionName(resourceName)
}

func (r *GenieSpaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Genie Space in Databricks.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "The title of the Genie Space.",
			},
			"warehouse_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the SQL warehouse to associate with this space.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the Genie Space.",
			},
			"parent_path": schema.StringAttribute{
				Optional:    true,
				Description: "The parent folder path where the space will be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Genie Space (same as id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tables": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "A list of fully qualified table names in catalog.schema.table format.",
			},
			"sample_questions": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "A list of sample question strings.",
			},
		},
	}
}

func (r *GenieSpaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.Client == nil && req.ProviderData != nil {
		r.Client = pluginfwcommon.ConfigureResource(req, resp)
	}
}

func (r *GenieSpaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var plan GenieSpaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.Client.WorkspaceClient()
	if err != nil {
		resp.Diagnostics.AddError("failed to get workspace client", err.Error())
		return
	}

	serializedSpace, diags := buildSerializedSpace(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := dashboards.GenieCreateSpaceRequest{
		Title:           plan.Title.ValueString(),
		WarehouseId:     plan.WarehouseId.ValueString(),
		Description:     plan.Description.ValueString(),
		ParentPath:      plan.ParentPath.ValueString(),
		SerializedSpace: serializedSpace,
	}

	space, err := w.Genie.CreateSpace(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create genie space", err.Error())
		return
	}

	plan.ID = types.StringValue(space.SpaceId)
	plan.SpaceId = types.StringValue(space.SpaceId)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GenieSpaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var state GenieSpaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.Client.WorkspaceClient()
	if err != nil {
		resp.Diagnostics.AddError("failed to get workspace client", err.Error())
		return
	}

	space, err := w.Genie.GetSpace(ctx, dashboards.GenieGetSpaceRequest{
		SpaceId:                state.ID.ValueString(),
		IncludeSerializedSpace: true,
		ForceSendFields:        []string{"IncludeSerializedSpace"},
	})
	if err != nil {
		if apierr.IsMissing(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to read genie space", err.Error())
		return
	}

	state.Title = types.StringValue(space.Title)
	state.WarehouseId = types.StringValue(space.WarehouseId)
	state.Description = types.StringValue(space.Description)
	state.SpaceId = types.StringValue(space.SpaceId)

	diags := parseSerializedSpace(ctx, &state, space.SerializedSpace)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *GenieSpaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var plan GenieSpaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the ID from state since it's computed
	var state GenieSpaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	plan.SpaceId = state.SpaceId

	w, err := r.Client.WorkspaceClient()
	if err != nil {
		resp.Diagnostics.AddError("failed to get workspace client", err.Error())
		return
	}

	serializedSpace, diags := buildSerializedSpace(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err = w.Genie.UpdateSpace(ctx, dashboards.GenieUpdateSpaceRequest{
		SpaceId:         state.ID.ValueString(),
		Title:           plan.Title.ValueString(),
		WarehouseId:     plan.WarehouseId.ValueString(),
		Description:     plan.Description.ValueString(),
		SerializedSpace: serializedSpace,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to update genie space", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GenieSpaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = pluginfwcontext.SetUserAgentInResourceContext(ctx, resourceName)

	var state GenieSpaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.Client.WorkspaceClient()
	if err != nil {
		resp.Diagnostics.AddError("failed to get workspace client", err.Error())
		return
	}

	err = w.Genie.TrashSpace(ctx, dashboards.GenieTrashSpaceRequest{
		SpaceId: state.ID.ValueString(),
	})
	if err != nil && !apierr.IsMissing(err) {
		resp.Diagnostics.AddError("failed to delete genie space", err.Error())
	}
}

func (r *GenieSpaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildSerializedSpace(ctx context.Context, model GenieSpaceModel) (string, diag.Diagnostics) {
	content := serializedSpaceContent{Version: 1}

	var tables []string
	if !model.Tables.IsNull() && !model.Tables.IsUnknown() {
		diags := model.Tables.ElementsAs(ctx, &tables, false)
		if diags.HasError() {
			return "", diags
		}
	}
	for _, t := range tables {
		content.DataSources.Tables = append(content.DataSources.Tables, serializedTable{
			Identifier: t,
		})
	}

	var questions []string
	if !model.SampleQuestions.IsNull() && !model.SampleQuestions.IsUnknown() {
		diags := model.SampleQuestions.ElementsAs(ctx, &questions, false)
		if diags.HasError() {
			return "", diags
		}
	}
	for _, q := range questions {
		content.Config.SampleQuestions = append(content.Config.SampleQuestions, serializedSampleQuestion{
			ID:       strings.ReplaceAll(uuid.New().String(), "-", ""),
			Question: []string{q},
		})
	}

	b, err := json.Marshal(content)
	if err != nil {
		return "", diag.Diagnostics{diag.NewErrorDiagnostic("failed to marshal serialized space", err.Error())}
	}
	return string(b), nil
}

func parseSerializedSpace(ctx context.Context, model *GenieSpaceModel, raw string) diag.Diagnostics {
	if raw == "" {
		return nil
	}
	var content serializedSpaceContent
	if err := json.Unmarshal([]byte(raw), &content); err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("failed to parse serialized space", err.Error())}
	}

	tables := make([]string, len(content.DataSources.Tables))
	for i, t := range content.DataSources.Tables {
		tables[i] = t.Identifier
	}
	tablesList, diags := types.ListValueFrom(ctx, types.StringType, tables)
	if diags.HasError() {
		return diags
	}
	model.Tables = tablesList

	questions := make([]string, 0, len(content.Config.SampleQuestions))
	for _, q := range content.Config.SampleQuestions {
		if len(q.Question) > 0 {
			questions = append(questions, q.Question[0])
		}
	}
	questionsList, diags := types.ListValueFrom(ctx, types.StringType, questions)
	if diags.HasError() {
		return diags
	}
	model.SampleQuestions = questionsList

	return nil
}
