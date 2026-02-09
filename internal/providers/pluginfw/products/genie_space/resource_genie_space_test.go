package genie_space

import (
	"context"
	"testing"

	"github.com/databricks/databricks-sdk-go/service/dashboards"
	"github.com/databricks/terraform-provider-databricks/internal/providers/pluginfw/converters"
	"github.com/databricks/terraform-provider-databricks/internal/service/dashboards_tf"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenieSpaceConversion(t *testing.T) {
	ctx := context.Background()

	// Test TF to SDK conversion
	tfSpace := dashboards_tf.GenieSpace{
		SpaceId:         types.StringValue("test-space-123"),
		Title:           types.StringValue("Test Space"),
		Description:     types.StringValue("Test description"),
		WarehouseId:     types.StringValue("warehouse-id"),
		SerializedSpace: types.StringValue(`{"version": "1.0"}`),
	}

	var sdkSpace dashboards.GenieSpace
	diags := converters.TfSdkToGoSdkStruct(ctx, tfSpace, &sdkSpace)
	require.False(t, diags.HasError())

	assert.Equal(t, "test-space-123", sdkSpace.SpaceId)
	assert.Equal(t, "Test Space", sdkSpace.Title)
	assert.Equal(t, "Test description", sdkSpace.Description)
	assert.Equal(t, "warehouse-id", sdkSpace.WarehouseId)
	assert.Equal(t, `{"version": "1.0"}`, sdkSpace.SerializedSpace)

	// Test SDK to TF conversion
	var tfSpaceResult dashboards_tf.GenieSpace
	diags = converters.GoSdkToTfSdkStruct(ctx, sdkSpace, &tfSpaceResult)
	require.False(t, diags.HasError())

	assert.Equal(t, "test-space-123", tfSpaceResult.SpaceId.ValueString())
	assert.Equal(t, "Test Space", tfSpaceResult.Title.ValueString())
	assert.Equal(t, "Test description", tfSpaceResult.Description.ValueString())
	assert.Equal(t, "warehouse-id", tfSpaceResult.WarehouseId.ValueString())
	assert.Equal(t, `{"version": "1.0"}`, tfSpaceResult.SerializedSpace.ValueString())
}

func TestGenieSpaceCreateRequestConversion(t *testing.T) {
	ctx := context.Background()

	// Test TF to SDK conversion for create request
	tfSpace := dashboards_tf.GenieSpace{
		Title:           types.StringValue("New Space"),
		Description:     types.StringValue("New description"),
		WarehouseId:     types.StringValue("warehouse-id"),
		SerializedSpace: types.StringValue(`{"version": "1.0", "config": {}}`),
	}

	var createRequest dashboards.GenieCreateSpaceRequest
	diags := converters.TfSdkToGoSdkStruct(ctx, tfSpace, &createRequest)
	require.False(t, diags.HasError())

	assert.Equal(t, "New Space", createRequest.Title)
	assert.Equal(t, "New description", createRequest.Description)
	assert.Equal(t, "warehouse-id", createRequest.WarehouseId)
	assert.Equal(t, `{"version": "1.0", "config": {}}`, createRequest.SerializedSpace)
}
