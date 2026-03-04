# NEXT CHANGELOG

## Release v1.111.0

### Breaking Changes

### New Features and Improvements

### Bug Fixes

* Fixed `lifecycle { ignore_changes }` not working for `spark_conf`, `spark_env_vars`, `custom_tags`, and `use_ml_runtime` on `databricks_cluster`. Externally-set values (via cluster policies or UI) were silently wiped on update ([#1238](https://github.com/databricks/terraform-provider-databricks/issues/1238)).

### Documentation

* Added documentation note about whitespace handling in `MAP` column types for `databricks_sql_table`.

### Exporter

### Internal Changes

* Host-agnostic cloud detection via node type patterns, replacing host-URL-based `IsAws()`/`IsAzure()`/`IsGcp()` checks.
