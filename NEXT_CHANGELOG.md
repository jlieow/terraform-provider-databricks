# NEXT CHANGELOG

## Release v1.115.0

### Breaking Changes

### New Features and Improvements

* Added `principal_id` argument to `databricks_git_credential` resource, enabling management of Git credentials on behalf of service principals.
* Add resource and data source for `databricks_postgres_catalog`.
* Add resource and data source for `databricks_postgres_synced_table`.
* Add resource and data sources for `databricks_environments_workspace_base_environment`.
* Add resource and data source for `databricks_environments_default_workspace_base_environment`.

* Added optional `cloud` argument to `databricks_current_config` data source to explicitly set the cloud type (`aws`, `azure`, `gcp`) instead of relying on host-based detection.

* Added `api` field to dual account/workspace resources (`databricks_user`, `databricks_service_principal`, `databricks_group`, `databricks_group_role`, `databricks_group_member`, `databricks_user_role`, `databricks_service_principal_role`, `databricks_user_instance_profile`, `databricks_group_instance_profile`, `databricks_metastore`, `databricks_metastore_assignment`, `databricks_metastore_data_access`, `databricks_storage_credential`, `databricks_service_principal_secret`, `databricks_access_control_rule_set`) to explicitly control whether account-level or workspace-level APIs are used. This enables support for unified hosts like `api.databricks.com` where the API level cannot be inferred from the host ([#5483](https://github.com/databricks/terraform-provider-databricks/pull/5483)).

### Bug Fixes

* Fix `databricks_service_principal` data source failing on account-level provider with `cannot populate provider_config for service principal: failed to resolve workspace_id` ([#5664](https://github.com/databricks/terraform-provider-databricks/issues/5664)). The data source now supports the `api` field and skips workspace-tracking when used at account level.
* Fix `databricks_service_principals` data source failing on account-level provider with the same `cannot populate provider_config for service principals: failed to resolve workspace_id` regression ([#5664](https://github.com/databricks/terraform-provider-databricks/issues/5664)). The data source now supports the `api` field and skips workspace-tracking when used at account level.
* Remove invalid `provider_config` attribute from account-only data sources `databricks_mws_workspaces` and `databricks_mws_credentials` ([#5664](https://github.com/databricks/terraform-provider-databricks/issues/5664)).

### Documentation

### Exporter

### Internal Changes
