package genie_space_test

import (
	"testing"

	"github.com/databricks/terraform-provider-databricks/internal/acceptance"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

const templateDataSource = baseTemplate + `
	resource "databricks_genie_space" "this" {
		title        = "Genie-DS-{var.STICKY_RANDOM}"
		warehouse_id = databricks_sql_endpoint.this.id
		description  = "Data source test"
		tables = [
			"samples.tpch.customer",
			"samples.tpch.lineitem",
			"samples.tpch.nation",
			"samples.tpch.orders",
			"samples.tpch.part",
			"samples.tpch.partsupp",
			"samples.tpch.region",
			"samples.tpch.supplier",
		]
		sample_questions = [
			"Show total revenue by month",
			"What are the top 10 customers?",
		]
	}

	data "databricks_genie_space" "this" {
		space_id = databricks_genie_space.this.id
	}
`

func TestAccGenieSpaceDataSource(t *testing.T) {
	acceptance.LoadWorkspaceEnv(t)
	acceptance.WorkspaceLevel(t, acceptance.Step{
		Template: templateDataSource,
		Check: func(s *terraform.State) error {
			r := s.RootModule().Resources["data.databricks_genie_space.this"]
			assert.Contains(t, r.Primary.Attributes["title"], "Genie-DS-")
			assert.Equal(t, "Data source test", r.Primary.Attributes["description"])
			assert.Equal(t, "8", r.Primary.Attributes["tables.#"])
			assert.Equal(t, "2", r.Primary.Attributes["sample_questions.#"])
			return nil
		},
	})
}
