# Fix: databricks_app Resource git_repository Update Issue

## Problem

When updating the `git_repository` field in a `databricks_app` resource, Terraform fails with:

```
Error: failed to update app

Git repository cannot be updated using update. Please use create-update instead.
```

## Root Cause

The `Update` method in `internal/providers/pluginfw/products/app/resource_app.go:285` uses the **synchronous** `Apps.Update()` API:

```go
response, err := w.Apps.Update(ctx, apps.UpdateAppRequest{App: appGoSdk, Name: app.Name.ValueString()})
```

However, the Databricks Apps API requires certain fields (including `git_repository`) to use the **asynchronous** `Apps.CreateUpdate()` API with an `update_mask` parameter.

## Fields Requiring CreateUpdate API

Based on the `AppUpdate` struct in the SDK (`apps/model.go:1232-1250`), these fields require `CreateUpdate`:

- `budget_policy_id`
- `compute_size`
- `description`
- **`git_repository`** ⬅️ Primary issue
- `resources`
- `usage_policy_id`
- `user_api_scopes`

## API Details

### Current (Broken) Approach
```go
// Uses synchronous Update API
w.Apps.Update(ctx, apps.UpdateAppRequest{
    App: appGoSdk,
    Name: app.Name.ValueString(),
})
```

### Required Approach
```go
// Use asynchronous CreateUpdate API
waiter, err := w.Apps.CreateUpdate(ctx, apps.AsyncUpdateAppRequest{
    AppName: app.Name.ValueString(),
    App: &appGoSdk,
    UpdateMask: "git_repository,description,resources", // Specify fields being updated
})
if err != nil {
    return err
}

// Wait for async update to complete
update, err := waiter.Poll(20*time.Minute, nil)
if err != nil {
    return err
}

// Then fetch the updated app
response, err := w.Apps.GetByName(ctx, app.Name.ValueString())
```

## Implementation Strategy

### Option 1: Always Use CreateUpdate (Recommended)
Replace the `Update` method to always use `CreateUpdate` with proper field masking. This ensures all updates work consistently.

**Pros:**
- Consistent behavior
- Future-proof for new fields
- Simpler logic

**Cons:**
- Slower updates (asynchronous)
- More complex implementation

### Option 2: Conditional Logic
Detect which fields changed and use `CreateUpdate` only when necessary.

**Pros:**
- Faster updates for simple fields
- Minimal changes to existing code

**Cons:**
- Complex field detection logic
- Risk of missing fields that require CreateUpdate
- Maintenance burden

## Recommended Solution

**Use Option 1**: Always use `CreateUpdate` API in the `Update` method.

### Implementation Steps

1. **Modify `Update` method** (`resource_app.go:258-304`):
   - Replace `w.Apps.Update()` with `w.Apps.CreateUpdate()`
   - Build the `update_mask` parameter by detecting changed fields
   - Wait for the async update to complete
   - Fetch the updated app state

2. **Build update_mask dynamically**:
   ```go
   // Compare plan vs state to determine changed fields
   changedFields := []string{}
   if !plan.Description.Equal(state.Description) {
       changedFields = append(changedFields, "description")
   }
   if !plan.GitRepository.Equal(state.GitRepository) {
       changedFields = append(changedFields, "git_repository")
   }
   // ... check other fields

   updateMask := strings.Join(changedFields, ",")
   ```

3. **Wait for completion** (similar to Create method's `waitForApp`):
   ```go
   waiter, err := w.Apps.CreateUpdate(ctx, apps.AsyncUpdateAppRequest{
       AppName: app.Name.ValueString(),
       App: &appGoSdk,
       UpdateMask: updateMask,
   })

   // Poll for completion
   update, err := waiter.Poll(20*time.Minute, nil)

   // Verify success
   if update.Status.State != apps.AppUpdateUpdateStatusUpdateStateSucceeded {
       return fmt.Errorf("update failed: %s", update.Status.Message)
   }
   ```

4. **Handle errors gracefully**:
   - Check for update failures
   - Provide clear error messages
   - Consider rollback scenarios

## Testing Requirements

1. **Unit Tests** (if possible with mocking):
   - Test updating `git_repository` alone
   - Test updating multiple fields together
   - Test update failures

2. **Acceptance Tests**:
   - Create app without git_repository, then add it
   - Update git_repository URL
   - Update git_repository provider
   - Update other fields alongside git_repository

3. **Manual Testing**:
   - Test with real GitHub repository
   - Test with private repositories (requires git credentials)
   - Test with different git providers

## Files to Modify

1. **`internal/providers/pluginfw/products/app/resource_app.go`**
   - Update `Update` method (lines 258-304)
   - Consider adding `buildUpdateMask` helper function
   - Reuse `waitForApp` or create similar waiting logic

2. **`docs/resources/app.md`**
   - Document `git_repository` field (currently missing)
   - Add example showing git_repository usage
   - Note that updates are asynchronous

## Documentation Gap

The `git_repository` field is **NOT documented** in `docs/resources/app.md` but exists in the schema. Add:

```markdown
* `git_repository` - (Optional) Git repository configuration for app deployments. When specified, deployments can reference code from this repository by providing only the git reference (branch, tag, or commit).
  * `provider` - (Required) Git provider. Supported values: `gitHub`, `gitHubEnterprise`, `bitbucketCloud`, `bitbucketServer`, `azureDevOpsServices`, `gitLab`, `gitLabEnterpriseEdition`, `awsCodeCommit`.
  * `url` - (Required) URL of the Git repository.
```

## Related Work

This fix is separate from the `databricks_app_deployment` Git source support being implemented in the `feature/resource_databricks_app_deploy` branch, but they complement each other:

- **This fix**: Allows configuring git_repository at the App level
- **Other feature**: Allows deploying from Git using that configuration

## References

- SDK AppUpdate struct: `apps/model.go:1232-1250`
- SDK AsyncUpdateAppRequest: `apps/model.go:1379-1395`
- Apps API docs: https://docs.databricks.com/api/workspace/apps/createupdate
