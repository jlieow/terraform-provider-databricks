package app

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestParseAppDeploymentId(t *testing.T) {
	state := AppDeploymentModel{
		AppName:      types.StringValue("my-app"),
		DeploymentId: types.StringValue("dep-123"),
	}
	appName, deploymentId, err := parseAppDeploymentId(state)
	assert.NoError(t, err)
	assert.Equal(t, "my-app", appName)
	assert.Equal(t, "dep-123", deploymentId)
}

func TestParseAppDeploymentId_MissingAppName(t *testing.T) {
	state := AppDeploymentModel{
		AppName:      types.StringValue(""),
		DeploymentId: types.StringValue("dep-123"),
	}
	_, _, err := parseAppDeploymentId(state)
	assert.Error(t, err)
}

func TestParseAppDeploymentId_MissingDeploymentId(t *testing.T) {
	state := AppDeploymentModel{
		AppName:      types.StringValue("my-app"),
		DeploymentId: types.StringValue(""),
	}
	_, _, err := parseAppDeploymentId(state)
	assert.Error(t, err)
}
