package app

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitSourceAttrTypes(t *testing.T) {
	attrTypes := gitSourceAttrTypes()

	assert.Equal(t, types.StringType, attrTypes["branch"])
	assert.Equal(t, types.StringType, attrTypes["tag"])
	assert.Equal(t, types.StringType, attrTypes["commit"])
	assert.Equal(t, types.StringType, attrTypes["source_code_path"])
	assert.Equal(t, types.StringType, attrTypes["resolved_commit"])

	gitRepoType, ok := attrTypes["git_repository"].(types.ObjectType)
	require.True(t, ok)
	assert.Equal(t, types.StringType, gitRepoType.AttrTypes["provider"])
	assert.Equal(t, types.StringType, gitRepoType.AttrTypes["url"])
}

func TestGitRepositoryAttrTypes(t *testing.T) {
	attrTypes := gitRepositoryAttrTypes()

	assert.Equal(t, types.StringType, attrTypes["provider"])
	assert.Equal(t, types.StringType, attrTypes["url"])
}
