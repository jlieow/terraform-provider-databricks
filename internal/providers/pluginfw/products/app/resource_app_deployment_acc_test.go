package app_test

import (
	"testing"

	"github.com/databricks/terraform-provider-databricks/internal/acceptance"
)

func TestAccAppDeploymentResource(t *testing.T) {
	acceptance.LoadWorkspaceEnv(t)
	if acceptance.IsGcp(t) {
		acceptance.Skipf(t)("not available on GCP")
	}
	acceptance.WorkspaceLevel(t, acceptance.Step{
		Template: `
	resource "databricks_app" "this" {
		name = "tf-{var.STICKY_RANDOM}"
		description = "app for deployment test"
		no_compute = true
	}

	resource "databricks_app_deployment" "this" {
		app_name         = databricks_app.this.name
		source_code_path = databricks_app.this.default_source_code_path
		triggers = {
			timestamp = timestamp()
		}
	}
		`,
	})
}
