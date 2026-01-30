---
subcategory: "Dashboards"
---

# databricks_genie_space Resource

This resource manages [Genie Spaces](https://docs.databricks.com/en/genie/index.html) in Databricks. Genie provides a no-code experience for business users, powered by AI/BI. Analysts set up spaces that business users can use to ask questions using natural language.

## Example Usage

```hcl
resource "databricks_genie_space" "example" {
  title        = "Revenue Analysis"
  warehouse_id = databricks_sql_endpoint.example.id
  description  = "Ask questions about revenue data"
  parent_path  = "/Workspace/Users/user@example.com"

  tables = [
    "catalog.schema.revenue",
    "catalog.schema.customers",
  ]

  sample_questions = [
    "Show total revenue by month",
    "What are the top 10 customers by revenue?",
  ]
}
```

## Argument Reference

The following arguments are required:

* `title` - (Required) The title of the Genie Space.
* `warehouse_id` - (Required) The ID of the SQL warehouse to associate with this space.
* `tables` - (Required) A list of fully qualified table names in `catalog.schema.table` format.

The following arguments are optional:

* `description` - (Optional) A description of the Genie Space.
* `parent_path` - (Optional, ForceNew) The parent folder path where the space will be created.
* `sample_questions` - (Optional) A list of sample question strings.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Genie Space.
* `space_id` - The ID of the Genie Space (same as `id`).

## Import

Genie Space can be imported using the space ID:

```bash
terraform import databricks_genie_space.example <space-id>
```
