# NEXT CHANGELOG

## Release v1.106.0

### Breaking Changes

### New Features and Improvements

* Add `role_arn` field to `databricks_mws_storage_configurations` resource to support sharing S3 buckets between root storage and Unity Catalog ([#5222](https://github.com/databricks/terraform-provider-databricks/issues/5222))
* Added support for updating `git_repository` on `databricks_app` resource and switched updates for `description`, `resources`, `budget_policy_id`, `compute_size`, `usage_policy_id`, and `user_api_scopes` to use the async `CreateUpdate` API, as the previous synchronous `Update` API does not support `git_repository` changes and does not handle fields that require async processing ([#XXXX](https://github.com/databricks/terraform-provider-databricks/pull/XXXX))

### Bug Fixes

* [Fix] `databricks_app` resource fail to read app when deleted outside terraform ([#5365](https://github.com/databricks/terraform-provider-databricks/pull/5365))

### Documentation

### Exporter

### Internal Changes
