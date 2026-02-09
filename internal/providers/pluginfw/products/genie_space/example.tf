# Example Terraform configuration for databricks_genie_space resource

resource "databricks_sql_endpoint" "this" {
  name             = "Genie Warehouse"
  cluster_size     = "Small"
  max_num_clusters = 1
}

resource "databricks_directory" "genie_dir" {
  path = "/Workspace/Shared/genie-spaces"
}

resource "databricks_genie_space" "example" {
  title       = "Revenue Analysis Space"
  description = "Ask questions about revenue data using TPC-H sample dataset"

  # Required: Warehouse to use for queries
  warehouse_id = databricks_sql_endpoint.this.id

  # Optional: Parent directory path where the space will be registered
  # parent_path is not supported in the create/update model - comment out if not working

  # Required: Serialized space configuration (JSON string)
  serialized_space = jsonencode({
    version = "1.0"
    config = {
      # Add your space configuration here
    }
  })
}

# Output the space ID
output "genie_space_id" {
  value = databricks_genie_space.example.space_id
}

# Import example:
# terraform import databricks_genie_space.example <space_id>
