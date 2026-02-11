---
subcategory: "Databricks SQL"
---

# databricks_sql_execution Resource

-> **Note** This resource executes arbitrary SQL statements using the [Statement Execution API](https://docs.databricks.com/api/workspace/statementexecution). It is intended for creating Databricks objects (such as metric views) that do not have dedicated REST APIs.

~> **Warning** This resource manages objects via SQL execution. Terraform cannot detect drift on the SQL content itself — only on whether the `read_sql` statement succeeds. If the object is modified outside of Terraform, those changes will not be detected.

## Example Usage

Creating a metric view:

```hcl
resource "databricks_sql_execution" "example_metric_view" {
  warehouse_id = databricks_sql_endpoint.example.id
  object_name  = "main.default.my_metric_view"

  create_sql = <<-SQL
    CREATE METRIC VIEW main.default.my_metric_view AS
    SELECT count(*) AS row_count
    FROM main.default.my_table
  SQL

  read_sql    = "DESCRIBE main.default.my_metric_view"
  destroy_sql = "DROP VIEW IF EXISTS main.default.my_metric_view"
}
```

## Argument Reference

The following arguments are required:

* `warehouse_id` - (Required) The ID of the SQL warehouse to use for executing the SQL statements.
* `create_sql` - (Required, Forces new resource) The SQL statement to execute when creating the resource.
* `read_sql` - (Required) The SQL statement to execute when checking if the object exists. If this statement succeeds, the object is considered to exist. If it fails, Terraform removes the resource from state.
* `destroy_sql` - (Required) The SQL statement to execute when destroying the resource.
* `object_name` - (Required, Forces new resource) The fully qualified name of the object being managed (e.g., `catalog.schema.object_name`). This is used as the resource ID.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The fully qualified name of the managed object (same as `object_name`).

## Import

The `databricks_sql_execution` resource can be imported using the `object_name`:

```bash
terraform import databricks_sql_execution.example "main.default.my_metric_view"
```

-> **Note** After import, you must ensure `create_sql`, `read_sql`, `destroy_sql`, and `warehouse_id` are configured in your Terraform configuration to match the imported object. Terraform will not be able to populate these fields automatically.
