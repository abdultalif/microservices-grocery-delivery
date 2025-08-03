package elasticsearch

import (
	"bytes"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/gommon/log"
)

func EnsureProductIndexExists(esClient *elasticsearch.Client) error {
	res, err := esClient.Indices.Exists([]string{"products"})
	if err != nil {
		return fmt.Errorf("error checking index: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		log.Info("Index 'products' already exists")
		return nil
	}

	if res.StatusCode == 404 {
		mapping := `{
			"mappings": {
				"properties": {
					"id": { "type": "keyword" },
					"category_slug": { "type": "keyword" },
					"parent_id": { "type": "keyword" },
					"name": { "type": "text" },
					"image": { "type": "text" },
					"description": { "type": "text" },
					"reguler_price": { "type": "double" },
					"sale_price": { "type": "double" },
					"unit": { "type": "keyword" },
					"weight": { "type": "integer" },
					"stock": { "type": "integer" },
					"variant": { "type": "keyword" },
					"status": { "type": "keyword" },
					"category_name": { "type": "text" },
					"category": {
						"type": "object",
						"properties": {
							"id": { "type": "keyword" },
							"parent_id": { "type": "keyword" },
							"name": { "type": "text" },
							"icon": { "type": "text" },
							"status": { "type": "boolean" },
							"slug": { "type": "keyword" },
							"description": { "type": "text" },
							"created_at": { "type": "date" },
							"updated_at": { "type": "date" }
						}
					},
					"created_at": { "type": "date" }
				}
			}
		}`

		createRes, err := esClient.Indices.Create("products", esClient.Indices.Create.WithBody(bytes.NewReader([]byte(mapping))))
		if err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}
		defer createRes.Body.Close()
		log.Info("Index 'products' created successfully")
		return nil
	}

	return fmt.Errorf("unexpected status code %d", res.StatusCode)
}


