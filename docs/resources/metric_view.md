---
subcategory: "Unity Catalog"
---
# databricks_metric_view (Resource)

This resource allows you to manage Databricks [Metric Views](https://docs.databricks.com/en/sql/language-manual/sql-ref-syntax-ddl-create-metric-view.html). A metric view is a reusable, governed definition of metrics stored in Unity Catalog. Metric views allow you to define metrics using a YAML specification and manage them as Unity Catalog objects.

-> This resource can only be used with a workspace-level provider!

## Example Usage

```hcl
data "databricks_sql_warehouse" "starter" {
  name = "Starter Warehouse"
}

resource "databricks_metric_view" "revenue" {
  name               = "revenue_metrics"
  catalog_name       = "main"
  schema_name        = "default"
  warehouse_id       = data.databricks_sql_warehouse.starter.id
  yaml_specification = <<-YAML
    source: samples.nyctaxi.trips
    measures:
      - name: total_fare
        type: INT
        expr: "fare_amount"
        agg: SUM
    dimensions:
      - name: pickup_zip
        type: INT
        expr: "pickup_zip"
    time_dimensions:
      - name: pickup_datetime
        type: TIMESTAMP
        expr: "tpep_pickup_datetime"
  YAML
}
```

## Argument Reference

The following arguments are required:

* `name` - The name of the metric view. Change forces creation of a new resource.
* `catalog_name` - The name of the catalog containing the metric view. Change forces creation of a new resource.
* `schema_name` - The name of the schema containing the metric view. Change forces creation of a new resource.
* `warehouse_id` - The ID of the SQL warehouse used to execute the SQL statements for managing the metric view.
* `yaml_specification` - The YAML specification defining the metric view.
* `provider_config` - (Optional) Configure the provider for management through account provider. This block consists of the following fields:
  * `workspace_id` - (Required) Workspace ID which the resource belongs to. This workspace must be part of the account which the provider is configured with.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The full name of the metric view in the form of `<catalog_name>.<schema_name>.<name>`.

## Import

This resource can be imported by its full three-level name: `<catalog_name>.<schema_name>.<name>`

Note: `warehouse_id` is not stored server-side and must be provided in the configuration after import.

```hcl
import {
  to = databricks_metric_view.this
  id = "<catalog_name>.<schema_name>.<name>"
}
```

Alternatively, when using `terraform` version 1.4 or earlier, import using the `terraform import` command:

```bash
terraform import databricks_metric_view.this <catalog_name>.<schema_name>.<name>
```
