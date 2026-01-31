package genie_space

import (
	"context"

	"github.com/databricks/databricks-sdk-go/service/dashboards"
	"github.com/databricks/terraform-provider-databricks/common"
	pluginfwcommon "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/common"
	pluginfwcontext "github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const dataSourceName = "genie_space"

var _ datasource.DataSourceWithConfigure = &GenieSpaceDataSource{}

func DataSourceGenieSpace() datasource.DataSource {
	return &GenieSpaceDataSource{}
}

type GenieSpaceDataSource struct {
	Client *common.DatabricksClient
}

type GenieSpaceDataModel struct {
	ID              types.String `tfsdk:"id"`
	SpaceId         types.String `tfsdk:"space_id"`
	Title           types.String `tfsdk:"title"`
	WarehouseId     types.String `tfsdk:"warehouse_id"`
	Description     types.String `tfsdk:"description"`
	Tables          types.List   `tfsdk:"tables"`
	SampleQuestions types.List   `tfsdk:"sample_questions"`
}

func (d *GenieSpaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = pluginfwcommon.GetDatabricksProductionName(dataSourceName)
}

func (d *GenieSpaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Genie Space from Databricks.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Genie Space.",
			},
			"space_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Genie Space to read.",
			},
			"title": schema.StringAttribute{
				Computed:    true,
				Description: "The title of the Genie Space.",
			},
			"warehouse_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the SQL warehouse associated with this space.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A description of the Genie Space.",
			},
			"tables": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A list of fully qualified table names in catalog.schema.table format.",
			},
			"sample_questions": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A list of sample question strings.",
			},
		},
	}
}

func (d *GenieSpaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if d.Client == nil && req.ProviderData != nil {
		d.Client = pluginfwcommon.ConfigureDataSource(req, resp)
	}
}

func (d *GenieSpaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = pluginfwcontext.SetUserAgentInDataSourceContext(ctx, dataSourceName)

	var config GenieSpaceDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := d.Client.WorkspaceClient()
	if err != nil {
		resp.Diagnostics.AddError("failed to get workspace client", err.Error())
		return
	}

	space, err := w.Genie.GetSpace(ctx, dashboards.GenieGetSpaceRequest{
		SpaceId:                config.SpaceId.ValueString(),
		IncludeSerializedSpace: true,
		ForceSendFields:        []string{"IncludeSerializedSpace"},
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to read genie space", err.Error())
		return
	}

	config.ID = types.StringValue(space.SpaceId)
	config.Title = types.StringValue(space.Title)
	config.WarehouseId = types.StringValue(space.WarehouseId)
	config.Description = types.StringValue(space.Description)

	// Reuse parseSerializedSpace with a temporary GenieSpaceModel
	model := &GenieSpaceModel{}
	diags := parseSerializedSpace(ctx, model, space.SerializedSpace)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Tables = model.Tables
	config.SampleQuestions = model.SampleQuestions

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
