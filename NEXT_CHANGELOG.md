# NEXT CHANGELOG

## Release v1.108.0

### Breaking Changes

### New Features and Improvements

### Bug Fixes

* Fixed `databricks_cluster` update silently wiping `spark_env_vars`, `spark_conf`, and `custom_tags` that were set outside of Terraform (e.g. by cluster policies) when those fields were not configured in HCL ([#1238](https://github.com/databricks/terraform-provider-databricks/issues/1238)).

### Documentation

### Exporter

### Internal Changes
