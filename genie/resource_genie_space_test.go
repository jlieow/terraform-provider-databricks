package genie

import (
	"testing"

	"github.com/databricks/databricks-sdk-go/experimental/mocks"
	"github.com/databricks/databricks-sdk-go/service/dashboards"
	"github.com/databricks/terraform-provider-databricks/qa"

	"github.com/stretchr/testify/mock"
)

func TestGenieSpaceCreate(t *testing.T) {
	qa.ResourceFixture{
		MockWorkspaceClientFunc: func(w *mocks.MockWorkspaceClient) {
			e := w.GetMockGenieAPI().EXPECT()
			e.CreateSpace(mock.Anything, mock.MatchedBy(func(req dashboards.GenieCreateSpaceRequest) bool {
				return req.Title == "My Genie Space" &&
					req.WarehouseId == "abc123" &&
					req.Description == "A test space" &&
					req.ParentPath == "/Workspace/Users/me" &&
					req.SerializedSpace != ""
			})).Return(&dashboards.GenieSpace{
				SpaceId:     "space-1",
				Title:       "My Genie Space",
				WarehouseId: "abc123",
				Description: "A test space",
			}, nil)
			e.GetSpace(mock.Anything, dashboards.GenieGetSpaceRequest{
				SpaceId:                "space-1",
				IncludeSerializedSpace: true,
				ForceSendFields:        []string{"IncludeSerializedSpace"},
			}).Return(&dashboards.GenieSpace{
				SpaceId:         "space-1",
				Title:           "My Genie Space",
				WarehouseId:     "abc123",
				Description:     "A test space",
				SerializedSpace: `{"version":1,"config":{"sample_questions":[{"id":"q1","question":["Show total revenue"]}]},"data_sources":{"tables":[{"identifier":"catalog.schema.table1"}]}}`,
			}, nil)
		},
		Resource: ResourceGenieSpace(),
		Create:   true,
		HCL: `
			title        = "My Genie Space"
			warehouse_id = "abc123"
			description  = "A test space"
			parent_path  = "/Workspace/Users/me"
			tables       = ["catalog.schema.table1"]
			sample_questions = ["Show total revenue"]
		`,
	}.ApplyAndExpectData(t, map[string]any{
		"id":       "space-1",
		"space_id": "space-1",
		"title":    "My Genie Space",
	})
}

func TestGenieSpaceRead(t *testing.T) {
	qa.ResourceFixture{
		MockWorkspaceClientFunc: func(w *mocks.MockWorkspaceClient) {
			w.GetMockGenieAPI().EXPECT().GetSpace(mock.Anything, dashboards.GenieGetSpaceRequest{
				SpaceId:                "space-1",
				IncludeSerializedSpace: true,
				ForceSendFields:        []string{"IncludeSerializedSpace"},
			}).Return(&dashboards.GenieSpace{
				SpaceId:         "space-1",
				Title:           "My Genie Space",
				WarehouseId:     "abc123",
				Description:     "A test space",
				SerializedSpace: `{"version":1,"config":{},"data_sources":{"tables":[{"identifier":"catalog.schema.table1"},{"identifier":"catalog.schema.table2"}]}}`,
			}, nil)
		},
		Resource: ResourceGenieSpace(),
		Read:     true,
		ID:       "space-1",
		HCL: `
			title        = "My Genie Space"
			warehouse_id = "abc123"
			description  = "A test space"
			tables       = ["catalog.schema.table1", "catalog.schema.table2"]
		`,
	}.ApplyAndExpectData(t, map[string]any{
		"id":    "space-1",
		"title": "My Genie Space",
	})
}

func TestGenieSpaceDelete(t *testing.T) {
	qa.ResourceFixture{
		MockWorkspaceClientFunc: func(w *mocks.MockWorkspaceClient) {
			w.GetMockGenieAPI().EXPECT().TrashSpace(mock.Anything, dashboards.GenieTrashSpaceRequest{
				SpaceId: "space-1",
			}).Return(nil)
		},
		Resource: ResourceGenieSpace(),
		Delete:   true,
		ID:       "space-1",
		HCL: `
			title        = "My Genie Space"
			warehouse_id = "abc123"
			tables       = ["catalog.schema.table1"]
		`,
	}.Apply(t)
}

func TestGenieSpaceUpdate(t *testing.T) {
	qa.ResourceFixture{
		MockWorkspaceClientFunc: func(w *mocks.MockWorkspaceClient) {
			e := w.GetMockGenieAPI().EXPECT()
			e.UpdateSpace(mock.Anything, mock.MatchedBy(func(req dashboards.GenieUpdateSpaceRequest) bool {
				return req.SpaceId == "space-1" &&
					req.Title == "Updated Title" &&
					req.WarehouseId == "abc123" &&
					req.SerializedSpace != ""
			})).Return(&dashboards.GenieSpace{
				SpaceId:     "space-1",
				Title:       "Updated Title",
				WarehouseId: "abc123",
			}, nil)
			e.GetSpace(mock.Anything, dashboards.GenieGetSpaceRequest{
				SpaceId:                "space-1",
				IncludeSerializedSpace: true,
				ForceSendFields:        []string{"IncludeSerializedSpace"},
			}).Return(&dashboards.GenieSpace{
				SpaceId:         "space-1",
				Title:           "Updated Title",
				WarehouseId:     "abc123",
				SerializedSpace: `{"version":1,"config":{},"data_sources":{"tables":[{"identifier":"catalog.schema.table1"}]}}`,
			}, nil)
		},
		Resource: ResourceGenieSpace(),
		Update:   true,
		ID:       "space-1",
		HCL: `
			title        = "Updated Title"
			warehouse_id = "abc123"
			tables       = ["catalog.schema.table1"]
		`,
	}.ApplyAndExpectData(t, map[string]any{
		"id":    "space-1",
		"title": "Updated Title",
	})
}
