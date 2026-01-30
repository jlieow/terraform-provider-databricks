package genie

import (
	"context"
	"encoding/json"

	"github.com/databricks/databricks-sdk-go/service/dashboards"
	"strings"

	"github.com/google/uuid"
	"github.com/databricks/terraform-provider-databricks/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type serializedSpaceContent struct {
	Version     int                          `json:"version"`
	Config      serializedSpaceConfig        `json:"config"`
	DataSources serializedSpaceDataSources   `json:"data_sources"`
}

type serializedSpaceConfig struct {
	SampleQuestions []serializedSampleQuestion `json:"sample_questions,omitempty"`
}

type serializedSpaceDataSources struct {
	Tables []serializedTable `json:"tables,omitempty"`
}

type serializedTable struct {
	Identifier string `json:"identifier"`
}

type serializedSampleQuestion struct {
	ID       string   `json:"id"`
	Question []string `json:"question"`
}

func ResourceGenieSpace() common.Resource {
	s := map[string]*schema.Schema{
		"title": {
			Type:     schema.TypeString,
			Required: true,
		},
		"warehouse_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"parent_path": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"tables": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"sample_questions": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"space_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}

	return common.Resource{
		Schema: s,
		Create: func(ctx context.Context, d *schema.ResourceData, c *common.DatabricksClient) error {
			w, err := c.WorkspaceClient()
			if err != nil {
				return err
			}
			serializedSpace, err := buildSerializedSpace(d)
			if err != nil {
				return err
			}
			resp, err := w.Genie.CreateSpace(ctx, dashboards.GenieCreateSpaceRequest{
				Title:           d.Get("title").(string),
				WarehouseId:     d.Get("warehouse_id").(string),
				Description:     d.Get("description").(string),
				ParentPath:      d.Get("parent_path").(string),
				SerializedSpace: serializedSpace,
			})
			if err != nil {
				return err
			}
			d.SetId(resp.SpaceId)
			d.Set("space_id", resp.SpaceId)
			return nil
		},
		Read: func(ctx context.Context, d *schema.ResourceData, c *common.DatabricksClient) error {
			w, err := c.WorkspaceClient()
			if err != nil {
				return err
			}
			resp, err := w.Genie.GetSpace(ctx, dashboards.GenieGetSpaceRequest{
				SpaceId:                d.Id(),
				IncludeSerializedSpace: true,
				ForceSendFields:        []string{"IncludeSerializedSpace"},
			})
			if err != nil {
				return err
			}
			d.Set("title", resp.Title)
			d.Set("warehouse_id", resp.WarehouseId)
			d.Set("description", resp.Description)
			d.Set("space_id", resp.SpaceId)
			return parseSerializedSpace(d, resp.SerializedSpace)
		},
		Update: func(ctx context.Context, d *schema.ResourceData, c *common.DatabricksClient) error {
			w, err := c.WorkspaceClient()
			if err != nil {
				return err
			}
			serializedSpace, err := buildSerializedSpace(d)
			if err != nil {
				return err
			}
			_, err = w.Genie.UpdateSpace(ctx, dashboards.GenieUpdateSpaceRequest{
				SpaceId:         d.Id(),
				Title:           d.Get("title").(string),
				WarehouseId:     d.Get("warehouse_id").(string),
				Description:     d.Get("description").(string),
				SerializedSpace: serializedSpace,
			})
			return err
		},
		Delete: func(ctx context.Context, d *schema.ResourceData, c *common.DatabricksClient) error {
			w, err := c.WorkspaceClient()
			if err != nil {
				return err
			}
			return w.Genie.TrashSpace(ctx, dashboards.GenieTrashSpaceRequest{
				SpaceId: d.Id(),
			})
		},
	}
}

func buildSerializedSpace(d *schema.ResourceData) (string, error) {
	content := serializedSpaceContent{Version: 1}
	for _, t := range d.Get("tables").([]interface{}) {
		content.DataSources.Tables = append(content.DataSources.Tables, serializedTable{
			Identifier: t.(string),
		})
	}
	for _, q := range d.Get("sample_questions").([]interface{}) {
		content.Config.SampleQuestions = append(content.Config.SampleQuestions, serializedSampleQuestion{
			ID:       strings.ReplaceAll(uuid.New().String(), "-", ""),
			Question: []string{q.(string)},
		})
	}
	b, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseSerializedSpace(d *schema.ResourceData, raw string) error {
	if raw == "" {
		return nil
	}
	var content serializedSpaceContent
	if err := json.Unmarshal([]byte(raw), &content); err != nil {
		return err
	}
	tables := make([]interface{}, len(content.DataSources.Tables))
	for i, t := range content.DataSources.Tables {
		tables[i] = t.Identifier
	}
	d.Set("tables", tables)
	questions := make([]interface{}, len(content.Config.SampleQuestions))
	for i, q := range content.Config.SampleQuestions {
		if len(q.Question) > 0 {
			questions[i] = q.Question[0]
		}
	}
	d.Set("sample_questions", questions)
	return nil
}
