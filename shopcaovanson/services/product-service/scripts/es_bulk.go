package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type esProductRow struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Price       float64   `db:"price"`
	SalePrice   *float64  `db:"sale_price"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	Category    string    `db:"category_slug"`
}

func bulkIndexElasticsearch(ctx context.Context, db *sqlx.DB, esURL, index string, limit int) error {
	if limit <= 0 {
		return nil
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{esURL}})
	if err != nil {
		return err
	}
	res, err := es.Ping()
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("elasticsearch ping failed")
	}

	ensureESIndex(es, index)

	const pageSize = 500
	indexed := 0
	for offset := 0; offset < limit; offset += pageSize {
		rows, err := fetchProductsForES(ctx, db, pageSize, offset)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		if err := esBulkIndex(es, index, rows); err != nil {
			return err
		}
		indexed += len(rows)
		log.Printf("elasticsearch indexed %d / %d products", indexed, limit)
	}
	return nil
}

func fetchProductsForES(ctx context.Context, db *sqlx.DB, limit, offset int) ([]esProductRow, error) {
	var rows []esProductRow
	err := db.SelectContext(ctx, &rows, `
		SELECT p.id, p.name, p.description, p.price, p.sale_price, p.is_active, p.created_at,
		       COALESCE(c.slug, '') AS category_slug
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		ORDER BY p.created_at
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return rows, err
}

func esBulkIndex(es *elasticsearch.Client, index string, rows []esProductRow) error {
	var buf bytes.Buffer
	for _, row := range rows {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
				"_id":    row.ID.String(),
			},
		}
		desc := ""
		if row.Description != nil {
			desc = *row.Description
		}
		doc := map[string]interface{}{
			"id":          row.ID.String(),
			"name":        row.Name,
			"description": desc,
			"category":    row.Category,
			"price":       row.Price,
			"is_active":   row.IsActive,
			"created_at":  row.CreatedAt.UTC().Format(time.RFC3339),
		}
		if row.SalePrice != nil {
			doc["sale_price"] = *row.SalePrice
		}
		metaLine, _ := json.Marshal(meta)
		docLine, _ := json.Marshal(doc)
		buf.Write(metaLine)
		buf.WriteByte('\n')
		buf.Write(docLine)
		buf.WriteByte('\n')
	}

	res, err := es.Bulk(bytes.NewReader(buf.Bytes()), es.Bulk.WithIndex(index))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("es bulk error: %s", string(b))
	}
	return nil
}

func ensureESIndex(es *elasticsearch.Client, index string) {
	res, err := es.Indices.Exists([]string{index})
	if err != nil || res == nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		return
	}
	mapping := `{
		"settings": {"number_of_shards": 1, "number_of_replicas": 0},
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"name": {"type": "text", "analyzer": "standard"},
				"description": {"type": "text"},
				"category": {"type": "keyword"},
				"price": {"type": "float"},
				"sale_price": {"type": "float"},
				"is_active": {"type": "boolean"},
				"created_at": {"type": "date"}
			}
		}
	}`
	cr, err := es.Indices.Create(index, es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		log.Printf("es create index warning: %v", err)
		return
	}
	cr.Body.Close()
}
