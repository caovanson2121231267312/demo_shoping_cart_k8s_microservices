package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
)

type SearchRepository interface {
	Available() bool
	Search(ctx context.Context, filter domain.ProductListFilter) ([]uuid.UUID, int, error)
}

type elasticsearchRepo struct {
	client *elasticsearch.Client
	index  string
}

func NewSearchRepository(client *elasticsearch.Client, index string) SearchRepository {
	return &elasticsearchRepo{client: client, index: index}
}

func (r *elasticsearchRepo) Available() bool {
	if r.client == nil {
		return false
	}
	res, err := r.client.Ping()
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return !res.IsError()
}

func (r *elasticsearchRepo) Search(ctx context.Context, filter domain.ProductListFilter) ([]uuid.UUID, int, error) {
	from := (filter.Page - 1) * filter.Limit

	must := []map[string]interface{}{}
	if filter.Search != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  filter.Search,
				"fields": []string{"name^2", "description"},
			},
		})
	}
	if filter.Category != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{"category": filter.Category},
		})
	}
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{"is_active": true},
	})

	rangeFilter := map[string]interface{}{}
	if filter.MinPrice != nil {
		rangeFilter["gte"] = *filter.MinPrice
	}
	if filter.MaxPrice != nil {
		rangeFilter["lte"] = *filter.MaxPrice
	}
	if len(rangeFilter) > 0 {
		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{"price": rangeFilter},
		})
	}

	sort := []map[string]interface{}{{"created_at": map[string]string{"order": "desc"}}}
	switch filter.Sort {
	case "price_asc":
		sort = []map[string]interface{}{{"price": map[string]string{"order": "asc"}}}
	case "price_desc":
		sort = []map[string]interface{}{{"price": map[string]string{"order": "desc"}}}
	case "newest":
		sort = []map[string]interface{}{{"created_at": map[string]string{"order": "desc"}}}
	}

	query := map[string]interface{}{
		"from": from,
		"size": filter.Limit,
		"sort": sort,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{"must": must},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, 0, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.index),
		r.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return nil, 0, fmt.Errorf("elasticsearch search error: %s", string(b))
	}

	var parsed struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source struct {
					ID string `json:"id"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, 0, err
	}

	ids := make([]uuid.UUID, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		id, err := uuid.Parse(strings.TrimSpace(hit.Source.ID))
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids, parsed.Hits.Total.Value, nil
}
