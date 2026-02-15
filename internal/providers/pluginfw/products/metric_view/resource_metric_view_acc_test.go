package metric_view_test

import (
	"testing"

	"github.com/databricks/terraform-provider-databricks/internal/acceptance"
)

var commonPartMetricView = `resource "databricks_catalog" "sandbox" {
	name          = "sandbox{var.STICKY_RANDOM}"
	comment       = "this catalog is managed by terraform"
	force_destroy = true
}

resource "databricks_schema" "things" {
	catalog_name = databricks_catalog.sandbox.id
	name         = "things{var.STICKY_RANDOM}"
	comment      = "this schema is managed by terraform"
}

resource "databricks_sql_table" "source" {
	catalog_name       = databricks_catalog.sandbox.id
	schema_name        = databricks_schema.things.name
	name               = "source{var.STICKY_RANDOM}"
	table_type         = "MANAGED"
	data_source_format = "DELTA"
	warehouse_id       = "{env.TEST_DEFAULT_WAREHOUSE_ID}"

	column {
		name = "revenue"
		type = "int"
	}
	column {
		name = "region"
		type = "string"
	}
	column {
		name = "ts"
		type = "timestamp"
	}
}
`

func TestUcAccMetricViewCreateAndUpdate(t *testing.T) {
	acceptance.UnityWorkspaceLevel(t, acceptance.Step{
		Template: commonPartMetricView + `
			resource "databricks_metric_view" "this" {
				name               = "mv{var.STICKY_RANDOM}"
				catalog_name       = databricks_catalog.sandbox.name
				schema_name        = databricks_schema.things.name
				warehouse_id       = "{env.TEST_DEFAULT_WAREHOUSE_ID}"
				yaml_specification = <<-YAML
source: ${databricks_sql_table.source.id}
measures:
  - name: total_revenue
    type: INT
    expr: "revenue"
    agg: SUM
dimensions:
  - name: region
    type: STRING
    expr: "region"
time_dimensions:
  - name: ts
    type: TIMESTAMP
    expr: "ts"
YAML
			}
		`,
	}, acceptance.Step{
		Template: commonPartMetricView + `
			resource "databricks_metric_view" "this" {
				name               = "mv{var.STICKY_RANDOM}"
				catalog_name       = databricks_catalog.sandbox.name
				schema_name        = databricks_schema.things.name
				warehouse_id       = "{env.TEST_DEFAULT_WAREHOUSE_ID}"
				yaml_specification = <<-YAML
source: ${databricks_sql_table.source.id}
measures:
  - name: total_revenue
    type: INT
    expr: "revenue"
    agg: SUM
  - name: avg_revenue
    type: INT
    expr: "revenue"
    agg: AVG
dimensions:
  - name: region
    type: STRING
    expr: "region"
time_dimensions:
  - name: ts
    type: TIMESTAMP
    expr: "ts"
YAML
			}
		`,
	})
}
