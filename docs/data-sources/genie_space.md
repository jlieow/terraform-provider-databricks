---
subcategory: "Dashboards"
---

# databricks_genie_space Data Source

Retrieves information about a [Genie Space](https://docs.databricks.com/en/genie/index.html) in Databricks.

## Example Usage

```hcl
data "databricks_genie_space" "example" {
  space_id = "01ab234567cd8901"
}
```

## Argument Reference

* `space_id` - (Required) The ID of the Genie Space to read.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Genie Space.
* `title` - The title of the Genie Space.
* `warehouse_id` - The ID of the SQL warehouse associated with this space.
* `description` - A description of the Genie Space.
* `tables` - A list of fully qualified table names in `catalog.schema.table` format.
* `sample_questions` - A list of sample question strings.
