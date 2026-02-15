package metric_view

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricViewResource_Metadata(t *testing.T) {
	r := ResourceMetricView()
	req := resource.MetadataRequest{ProviderTypeName: "databricks"}
	resp := resource.MetadataResponse{}
	r.Metadata(context.Background(), req, &resp)
	assert.Equal(t, "databricks_metric_view", resp.TypeName)
}

func TestMetricViewResource_Schema(t *testing.T) {
	r := ResourceMetricView()
	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError(), "schema generation should not produce errors")

	schema := resp.Schema

	for _, attr := range []string{"warehouse_id", "name", "catalog_name", "schema_name", "yaml_specification"} {
		a, ok := schema.Attributes[attr]
		assert.True(t, ok, "schema should contain attribute %s", attr)
		assert.True(t, a.IsRequired(), "attribute %s should be required", attr)
	}

	idAttr, ok := schema.Attributes["id"]
	assert.True(t, ok, "schema should contain attribute id")
	assert.True(t, idAttr.IsComputed(), "id should be computed")
	assert.True(t, idAttr.IsOptional(), "id should be optional")
}

func TestMetricViewModel_GetComplexFieldTypes(t *testing.T) {
	m := MetricViewModel{}
	fieldTypes := m.GetComplexFieldTypes(context.Background())
	_, ok := fieldTypes["provider_config"]
	assert.True(t, ok, "should declare provider_config as a complex field type")
}

func TestMetricViewResource_Interfaces(t *testing.T) {
	r := ResourceMetricView()
	var _ resource.ResourceWithConfigure = r.(*metricViewResource)
	var _ resource.ResourceWithImportState = r.(*metricViewResource)
}

func TestMetricViewModel_FullName(t *testing.T) {
	m := MetricViewModel{
		CatalogName: types.StringValue("main"),
		SchemaName:  types.StringValue("default"),
		Name:        types.StringValue("my_metric"),
	}
	assert.Equal(t, "main.default.my_metric", m.fullName())
}

func TestMetricViewModel_SQLFullName(t *testing.T) {
	m := MetricViewModel{
		CatalogName: types.StringValue("main"),
		SchemaName:  types.StringValue("default"),
		Name:        types.StringValue("my_metric"),
	}
	assert.Equal(t, "`main`.`default`.`my_metric`", m.sqlFullName())
}

func TestSuppressCaseChangePlanModifier(t *testing.T) {
	mod := suppressCaseChangePlanModifier{}

	t.Run("suppresses case-only change", func(t *testing.T) {
		req := planmodifier.StringRequest{
			StateValue: types.StringValue("Main"),
			PlanValue:  types.StringValue("main"),
		}
		resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
		mod.PlanModifyString(context.Background(), req, resp)
		assert.Equal(t, "Main", resp.PlanValue.ValueString())
	})

	t.Run("allows real change", func(t *testing.T) {
		req := planmodifier.StringRequest{
			StateValue: types.StringValue("main"),
			PlanValue:  types.StringValue("other"),
		}
		resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
		mod.PlanModifyString(context.Background(), req, resp)
		assert.Equal(t, "other", resp.PlanValue.ValueString())
	})

	t.Run("handles null state", func(t *testing.T) {
		req := planmodifier.StringRequest{
			StateValue: types.StringNull(),
			PlanValue:  types.StringValue("main"),
		}
		resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
		mod.PlanModifyString(context.Background(), req, resp)
		assert.Equal(t, "main", resp.PlanValue.ValueString())
	})
}

func TestMetricViewResource_ImportState_InvalidID(t *testing.T) {
	r := &metricViewResource{}
	req := resource.ImportStateRequest{ID: "invalid_id"}
	resp := resource.ImportStateResponse{}
	r.ImportState(context.Background(), req, &resp)
	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "expected format")
}

func TestMetricViewResource_ImportState_TooManyParts(t *testing.T) {
	r := &metricViewResource{}
	// SplitN with n=3 means "a.b.c.d" becomes ["a", "b", "c.d"] which is valid (3 parts)
	// but "invalid" with no dots becomes ["invalid"] which is invalid
	req := resource.ImportStateRequest{ID: "only_one_part"}
	resp := resource.ImportStateResponse{}
	r.ImportState(context.Background(), req, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}
