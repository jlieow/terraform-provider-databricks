---
subcategory: "Apps"
---

# databricks_app_deployment Resource

This resource triggers a deployment for a [databricks_app](app.md). Every configuration change forces the resource to be replaced, creating a new deployment.

This resource is useful for managing app deployments declaratively. Use the `triggers` argument to control when a new deployment is created.

## Example Usage

```hcl
resource "databricks_app" "this" {
  name        = "my-app"
  description = "My Databricks app"
}

resource "databricks_app_deployment" "this" {
  app_name         = databricks_app.this.name
  source_code_path = databricks_app.this.default_source_code_path
  triggers = {
    source_code_hash = filemd5("${path.module}/app/main.py")
  }
}
```

## Argument Reference

The following arguments are required:

* `app_name` - (Required, Forces new resource) The name of the app to deploy.
* `source_code_path` - (Required, Forces new resource) The workspace file system path of the source code used for the deployment.

The following arguments are optional:

* `triggers` - (Optional, Forces new resource) Arbitrary map of string values that, when changed, will trigger a new deployment. Use this to redeploy when source code or configuration changes.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `deployment_id` - The unique ID of the deployment.

## Import

This resource can be imported using the app name and deployment ID separated by a pipe character:

```bash
terraform import databricks_app_deployment.this "my-app|deployment-id"
```
