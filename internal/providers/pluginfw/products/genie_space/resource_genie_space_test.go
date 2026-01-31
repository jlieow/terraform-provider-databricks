package genie_space

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSerializedSpace(t *testing.T) {
	ctx := context.Background()
	tablesList, _ := types.ListValueFrom(ctx, types.StringType, []string{"catalog.schema.table1", "catalog.schema.table2"})
	questionsList, _ := types.ListValueFrom(ctx, types.StringType, []string{"How many rows?", "Show total"})

	model := GenieSpaceModel{
		Title:           types.StringValue("test"),
		WarehouseId:     types.StringValue("wh-1"),
		Tables:          tablesList,
		SampleQuestions: questionsList,
	}

	result, diags := buildSerializedSpace(ctx, model)
	require.False(t, diags.HasError())
	assert.Contains(t, result, `"identifier":"catalog.schema.table1"`)
	assert.Contains(t, result, `"identifier":"catalog.schema.table2"`)
	assert.Contains(t, result, `"How many rows?"`)
	assert.Contains(t, result, `"Show total"`)
	assert.Contains(t, result, `"version":1`)
}

func TestParseSerializedSpace(t *testing.T) {
	ctx := context.Background()
	raw := `{"version":1,"config":{"sample_questions":[{"id":"q1","question":["Show total revenue"]},{"id":"q2","question":["Top customers"]}]},"data_sources":{"tables":[{"identifier":"catalog.schema.table1"}]}}`

	model := &GenieSpaceModel{}
	diags := parseSerializedSpace(ctx, model, raw)
	require.Nil(t, diags)

	var tables []string
	model.Tables.ElementsAs(ctx, &tables, false)
	assert.Equal(t, []string{"catalog.schema.table1"}, tables)

	var questions []string
	model.SampleQuestions.ElementsAs(ctx, &questions, false)
	assert.Equal(t, []string{"Show total revenue", "Top customers"}, questions)
}

func TestParseSerializedSpaceEmpty(t *testing.T) {
	ctx := context.Background()
	model := &GenieSpaceModel{}
	diags := parseSerializedSpace(ctx, model, "")
	assert.Nil(t, diags)
}
