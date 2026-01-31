package genie_space_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/dashboards"
	"github.com/databricks/terraform-provider-databricks/internal/acceptance"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

const baseTemplate = `
	resource "databricks_sql_endpoint" "this" {
		name             = "tf-genie-{var.STICKY_RANDOM}"
		cluster_size     = "2X-Small"
		max_num_clusters = 1
	}
`

const templateDriftTitle = baseTemplate + `
	resource "databricks_genie_space" "this" {
		title        = "Genie-{var.STICKY_RANDOM}"
		warehouse_id = databricks_sql_endpoint.this.id
		description  = "Drift test"
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
`

const templateDriftTable = baseTemplate + `
	resource "databricks_genie_space" "this" {
		title        = "Genie-Table-{var.STICKY_RANDOM}"
		warehouse_id = databricks_sql_endpoint.this.id
		description  = "Drift table test"
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
`

const templateDriftQuestion = baseTemplate + `
	resource "databricks_genie_space" "this" {
		title        = "Genie-Question-{var.STICKY_RANDOM}"
		warehouse_id = databricks_sql_endpoint.this.id
		description  = "Drift question test"
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
`

func TestAccGenieSpaceDriftTitle(t *testing.T) {
	var spaceID string
	acceptance.LoadWorkspaceEnv(t)
	acceptance.WorkspaceLevel(t, acceptance.Step{
		Template: templateDriftTitle,
		Check: func(s *terraform.State) error {
			r := s.RootModule().Resources["databricks_genie_space.this"]
			spaceID = r.Primary.ID
			return nil
		},
	}, acceptance.Step{
		PreConfig: func() {
			w := databricks.Must(databricks.NewWorkspaceClient())
			ctx := context.Background()
			_, err := w.Genie.UpdateSpace(ctx, dashboards.GenieUpdateSpaceRequest{
				SpaceId: spaceID,
				Title:   "Drifted Title",
			})
			require.NoError(t, err)
		},
		Template: templateDriftTitle,
		Check: func(s *terraform.State) error {
			r := s.RootModule().Resources["databricks_genie_space.this"]
			assert.Contains(t, r.Primary.Attributes["title"], "Genie-")
			assert.NotEqual(t, "Drifted Title", r.Primary.Attributes["title"])
			return nil
		},
	})
}

func TestAccGenieSpaceDriftTable(t *testing.T) {
	var spaceID string
	acceptance.LoadWorkspaceEnv(t)
	acceptance.WorkspaceLevel(t, acceptance.Step{
		Template: templateDriftTable,
		Check: func(s *terraform.State) error {
			r := s.RootModule().Resources["databricks_genie_space.this"]
			spaceID = r.Primary.ID
			return nil
		},
	}, acceptance.Step{
		PreConfig: func() {
			w := databricks.Must(databricks.NewWorkspaceClient())
			ctx := context.Background()
			space, err := w.Genie.GetSpace(ctx, dashboards.GenieGetSpaceRequest{
				SpaceId:                spaceID,
				IncludeSerializedSpace: true,
				ForceSendFields:        []string{"IncludeSerializedSpace"},
			})
			require.NoError(t, err)
			var content serializedSpaceContent
			require.NoError(t, json.Unmarshal([]byte(space.SerializedSpace), &content))
			// Remove all but the first table to create drift
			content.DataSources.Tables = content.DataSources.Tables[:1]
			b, err := json.Marshal(content)
			require.NoError(t, err)
			_, err = w.Genie.UpdateSpace(ctx, dashboards.GenieUpdateSpaceRequest{
				SpaceId:         spaceID,
				Title:           space.Title,
				WarehouseId:     space.WarehouseId,
				Description:     space.Description,
				SerializedSpace: string(b),
			})
			require.NoError(t, err)
		},
		Template: templateDriftTable,
		Check: func(s *terraform.State) error {
			r := s.RootModule().Resources["databricks_genie_space.this"]
			assert.Equal(t, "8", r.Primary.Attributes["tables.#"])
			assert.Equal(t, "samples.tpch.customer", r.Primary.Attributes["tables.0"])
			return nil
		},
	})
}

func TestAccGenieSpaceDriftSampleQuestion(t *testing.T) {
	var spaceID string
	acceptance.LoadWorkspaceEnv(t)
	acceptance.WorkspaceLevel(t, acceptance.Step{
		Template: templateDriftQuestion,
		Check: func(s *terraform.State) error {
			r := s.RootModule().Resources["databricks_genie_space.this"]
			spaceID = r.Primary.ID
			return nil
		},
	}, acceptance.Step{
		PreConfig: func() {
			w := databricks.Must(databricks.NewWorkspaceClient())
			ctx := context.Background()
			space, err := w.Genie.GetSpace(ctx, dashboards.GenieGetSpaceRequest{
				SpaceId:                spaceID,
				IncludeSerializedSpace: true,
				ForceSendFields:        []string{"IncludeSerializedSpace"},
			})
			require.NoError(t, err)
			var content serializedSpaceContent
			require.NoError(t, json.Unmarshal([]byte(space.SerializedSpace), &content))
			content.Config.SampleQuestions = nil
			b, err := json.Marshal(content)
			require.NoError(t, err)
			_, err = w.Genie.UpdateSpace(ctx, dashboards.GenieUpdateSpaceRequest{
				SpaceId:         spaceID,
				Title:           space.Title,
				WarehouseId:     space.WarehouseId,
				Description:     space.Description,
				SerializedSpace: string(b),
			})
			require.NoError(t, err)
		},
		Template: templateDriftQuestion,
		Check: func(s *terraform.State) error {
			r := s.RootModule().Resources["databricks_genie_space.this"]
			assert.Equal(t, "2", r.Primary.Attributes["sample_questions.#"])
			assert.Equal(t, "Show total revenue by month", r.Primary.Attributes["sample_questions.0"])
			assert.Equal(t, "What are the top 10 customers?", r.Primary.Attributes["sample_questions.1"])
			return nil
		},
	})
}
