# NEXT CHANGELOG

## Release v1.108.0

### Breaking Changes

### New Features and Improvements

* Added `database_project_name` to `databricks_permissions` for managing Lakebase database project permissions.
* Added `node_type_flexibility` block to `databricks_instance_pool` resource ([#5381](https://github.com/databricks/terraform-provider-databricks/pull/5381)).

### Bug Fixes

* Fixed `databricks_cluster` ignoring empty map values for `spark_env_vars`, `spark_conf`, and `custom_tags`, which prevented clearing previously-set values.
* Handle error during WorkspaceClient() creation in databricks_grant and databricks_grants resources ([#5403](https://github.com/databricks/terraform-provider-databricks/pull/5403))

### Documentation

* Added note to `databricks_mws_ncc_binding` that a workspace can only have one NCC binding at a time.

### Exporter

### Internal Changes
