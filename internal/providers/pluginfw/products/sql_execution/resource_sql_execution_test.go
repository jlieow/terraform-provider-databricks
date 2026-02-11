package sql_execution

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestSqlExecutionResource_Metadata(t *testing.T) {
	r := ResourceSqlExecution()
	req := resource.MetadataRequest{ProviderTypeName: "databricks"}
	resp := resource.MetadataResponse{}
	r.Metadata(context.Background(), req, &resp)
	assert.Equal(t, "databricks_sql_execution", resp.TypeName)
}

func TestSqlExecutionResource_Schema(t *testing.T) {
	r := ResourceSqlExecution()
	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError(), "schema generation should not produce errors")

	schema := resp.Schema

	// Verify required attributes
	for _, attr := range []string{"warehouse_id", "create_sql", "read_sql", "destroy_sql", "object_name"} {
		a, ok := schema.Attributes[attr]
		assert.True(t, ok, "schema should contain attribute %s", attr)
		assert.True(t, a.IsRequired(), "attribute %s should be required", attr)
	}

	// Verify id is computed + optional
	idAttr, ok := schema.Attributes["id"]
	assert.True(t, ok, "schema should contain attribute id")
	assert.True(t, idAttr.IsComputed(), "id should be computed")
	assert.True(t, idAttr.IsOptional(), "id should be optional")
}

func TestSqlExecutionModel_GetComplexFieldTypes(t *testing.T) {
	m := SqlExecutionModel{}
	types := m.GetComplexFieldTypes(context.Background())
	_, ok := types["provider_config"]
	assert.True(t, ok, "should declare provider_config as a complex field type")
}

func TestSqlExecutionResource_Interfaces(t *testing.T) {
	r := ResourceSqlExecution()
	var _ resource.ResourceWithConfigure = r.(*sqlExecutionResource)
	var _ resource.ResourceWithImportState = r.(*sqlExecutionResource)
}
