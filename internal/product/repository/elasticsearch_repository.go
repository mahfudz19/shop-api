package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/username/shop-api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ElasticsearchProductRepository implements ProductRepository with Elasticsearch backend.
type ElasticsearchProductRepository struct {
	client *elasticsearch.Client
	index  string
}

// NewElasticsearchProductRepository creates a new Elasticsearch product repository.
// Returns concrete type to allow access to ES-specific methods.
func NewElasticsearchProductRepository(client *elasticsearch.Client) *ElasticsearchProductRepository {
	return &ElasticsearchProductRepository{
		client: client,
		index:  "products",
	}
}

// FetchWithFilter fetches products with filter from Elasticsearch
func (e *ElasticsearchProductRepository) FetchWithFilter(
	ctx context.Context,
	filter domain.ProductFilter,
) ([]domain.Product, int64, error) {
	query, err := e.buildSearchQuery(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build query: %w", err)
	}

	req := &esapi.SearchRequest{
		Index: []string{e.index},
		Body:  query,
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return nil, 0, fmt.Errorf("search request failed: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("⚠️  Warning: Failed to close response body: %v", closeErr)
		}
	}()

	if res.IsError() {
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(res.Body); err != nil {
			return nil, 0, fmt.Errorf("failed to read error body: %w", err)
		}
		return nil, 0, fmt.Errorf("search error: %s", buf.String())
	}

	var result SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("decode response failed: %w", err)
	}

	products := make([]domain.Product, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		var product domain.Product
		if err := json.Unmarshal(hit.Source, &product); err != nil {
			log.Printf("⚠️  Warning: Failed to decode product: %v", err)
			continue
		}
		products = append(products, product)
	}

	return products, result.Hits.Total.Value, nil
}

// buildSearchQuery builds Elasticsearch query from filter
func (e *ElasticsearchProductRepository) buildSearchQuery(filter domain.ProductFilter) (*bytes.Buffer, error) {
	boolQuery := map[string]interface{}{
		"bool": map[string]interface{}{
			"must":   make([]map[string]interface{}, 0),
			"filter": make([]map[string]interface{}, 0),
		},
	}

	boolClause := boolQuery["bool"].(map[string]interface{})

	// Search query
	if filter.Search != "" {
		boolClause["must"] = append(boolClause["must"].([]map[string]interface{}),
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  filter.Search,
					"fields": []string{"name^2", "search_keyword^1.5", "category"},
					"type":   "best_fields",
				},
			},
		)
	}

	// Match all if no search term
	if filter.Search == "" {
		boolClause["must"] = append(boolClause["must"].([]map[string]interface{}),
			map[string]interface{}{"match_all": map[string]interface{}{}},
		)
	}

	// Filters
	if filter.Location != "" {
		boolClause["filter"] = append(boolClause["filter"].([]map[string]interface{}),
			map[string]interface{}{"term": map[string]interface{}{"location": filter.Location}},
		)
	}

	if filter.Marketplace != "" {
		boolClause["filter"] = append(boolClause["filter"].([]map[string]interface{}),
			map[string]interface{}{"term": map[string]interface{}{"marketplace": filter.Marketplace}},
		)
	}

	// Price range
	if filter.MinPrice > 0 || filter.MaxPrice > 0 {
		priceFilter := map[string]interface{}{}
		if filter.MinPrice > 0 {
			priceFilter["gte"] = filter.MinPrice
		}
		if filter.MaxPrice > 0 {
			priceFilter["lte"] = filter.MaxPrice
		}
		boolClause["filter"] = append(boolClause["filter"].([]map[string]interface{}),
			map[string]interface{}{"range": map[string]interface{}{"price_rp": priceFilter}},
		)
	}

	// Rating filter
	if filter.Rating > 0 {
		boolClause["filter"] = append(boolClause["filter"].([]map[string]interface{}),
			map[string]interface{}{"range": map[string]interface{}{"rating": map[string]interface{}{"gte": filter.Rating}}},
		)
	}

	// Build query body
	queryBody := map[string]interface{}{
		"query": boolQuery,
		"from":  (filter.Page - 1) * filter.Limit,
		"size":  filter.Limit,
	}

	// Sorting
	sortField := filter.SortBy
	if sortField == "" {
		sortField = "created_at"
	}

	sortOrder := "asc"
	if filter.SortOrder == "desc" {
		sortOrder = "desc"
	}

	// Map field names to snake_case
	esSortField := sortField
	switch sortField {
	case "createdAt":
		esSortField = "created_at"
	case "updatedAt":
		esSortField = "updated_at"
	case "priceRp":
		esSortField = "price_rp"
	}

	queryBody["sort"] = []map[string]interface{}{
		{esSortField: map[string]interface{}{"order": sortOrder}},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(queryBody); err != nil {
		return nil, err
	}

	return &buf, nil
}

// GetByID gets a product by ID from Elasticsearch
func (e *ElasticsearchProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	var product domain.Product

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return product, fmt.Errorf("invalid ID format: %w", err)
	}

	req := &esapi.GetRequest{
		Index:      e.index,
		DocumentID: objID.Hex(),
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return product, fmt.Errorf("get request failed: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("⚠️  Warning: Failed to close response body: %v", closeErr)
		}
	}()

	if res.StatusCode == 404 {
		return product, fmt.Errorf("product not found")
	}

	if res.IsError() {
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(res.Body); err != nil {
			return product, fmt.Errorf("failed to read error body: %w", err)
		}
		return product, fmt.Errorf("get error: %s", buf.String())
	}

	var result GetResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return product, fmt.Errorf("decode response failed: %w", err)
	}

	if !result.Found {
		return product, fmt.Errorf("product not found")
	}

	if err := json.Unmarshal(result.Source, &product); err != nil {
		return product, fmt.Errorf("decode product failed: %w", err)
	}

	return product, nil
}

// GetDeals gets products with highest discount
func (e *ElasticsearchProductRepository) GetDeals(ctx context.Context, limit int64) ([]domain.Product, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"discount_percent": map[string]interface{}{"gt": 0},
			},
		},
		"sort": []map[string]interface{}{
			{"discount_percent": map[string]interface{}{"order": "desc"}},
		},
		"size": limit,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	req := &esapi.SearchRequest{
		Index: []string{e.index},
		Body:  &buf,
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("⚠️  Warning: Failed to close response body: %v", closeErr)
		}
	}()

	if res.IsError() {
		var body bytes.Buffer
		if _, err := body.ReadFrom(res.Body); err != nil {
			return nil, fmt.Errorf("failed to read error body: %w", err)
		}
		return nil, fmt.Errorf("search error: %s", body.String())
	}

	var result SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	products := make([]domain.Product, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		var product domain.Product
		if err := json.Unmarshal(hit.Source, &product); err != nil {
			log.Printf("⚠️  Warning: Failed to decode product: %v", err)
			continue
		}
		products = append(products, product)
	}

	return products, nil
}

// GetStats gets product statistics using aggregations
func (e *ElasticsearchProductRepository) GetStats(ctx context.Context) (domain.ProductStats, error) {
	query := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"total_products": map[string]interface{}{
				"value_count": map[string]interface{}{"field": "id"},
			},
			"total_shops": map[string]interface{}{
				"cardinality": map[string]interface{}{"field": "shop"},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return domain.ProductStats{}, fmt.Errorf("failed to encode query: %w", err)
	}

	req := &esapi.SearchRequest{
		Index: []string{e.index},
		Body:  &buf,
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return domain.ProductStats{}, fmt.Errorf("search request failed: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("⚠️  Warning: Failed to close response body: %v", closeErr)
		}
	}()

	if res.IsError() {
		var body bytes.Buffer
		if _, err := body.ReadFrom(res.Body); err != nil {
			return domain.ProductStats{}, fmt.Errorf("failed to read error body: %w", err)
		}
		return domain.ProductStats{}, fmt.Errorf("search error: %s", body.String())
	}

	var result StatsResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return domain.ProductStats{}, fmt.Errorf("decode response failed: %w", err)
	}

	return domain.ProductStats{
		TotalProducts: int64(result.Aggregations.TotalProducts.Value),
		TotalShops:    int(result.Aggregations.TotalShops.Value),
	}, nil
}

// IndexProduct indexes a single product to Elasticsearch
func (e *ElasticsearchProductRepository) IndexProduct(ctx context.Context, product domain.Product) error {
	data, err := json.Marshal(e.toESDocument(product))
	if err != nil {
		return fmt.Errorf("marshal product failed: %w", err)
	}

	req := &esapi.IndexRequest{
		Index:      e.index,
		DocumentID: product.ID.Hex(),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("index request failed: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("⚠️  Warning: Failed to close response body: %v", closeErr)
		}
	}()

	if res.IsError() {
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(res.Body); err != nil {
			return fmt.Errorf("failed to read error body: %w", err)
		}
		return fmt.Errorf("index error: %s", buf.String())
	}

	return nil
}

// DeleteProduct deletes a product from Elasticsearch
func (e *ElasticsearchProductRepository) DeleteProduct(ctx context.Context, id string) error {
	req := &esapi.DeleteRequest{
		Index:      e.index,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("⚠️  Warning: Failed to close response body: %v", closeErr)
		}
	}()

	if res.IsError() && res.StatusCode != 404 {
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(res.Body); err != nil {
			return fmt.Errorf("failed to read error body: %w", err)
		}
		return fmt.Errorf("delete error: %s", buf.String())
	}

	return nil
}

// toESDocument converts Product to Elasticsearch document
func (e *ElasticsearchProductRepository) toESDocument(product domain.Product) map[string]interface{} {
	return map[string]interface{}{
		"id":                     product.ID.Hex(),
		"master_product_id":      product.MasterProductID.Hex(),
		"name":                   product.Name,
		"url":                    product.URL,
		"clean_url":              product.CleanURL,
		"category":               product.Category,
		"search_keyword":         product.SearchKeyword,
		"price_rp":               product.PriceRp,
		"price_original":         product.PriceOriginal,
		"discount_percent":       product.DiscountPercent,
		"rating":                 product.Rating,
		"sold_count":             product.SoldCount,
		"location":               product.Location,
		"marketplace":            product.Marketplace,
		"shop":                   product.Shop,
		"image_url":              product.ImageURL,
		"is_anomaly":             product.IsAnomaly,
		"match_confidence":       product.MatchConfidence,
		"marketplace_product_id": product.MarketplaceProductID,
		"created_at":             product.CreatedAt,
		"updated_at":             product.UpdatedAt,
	}
}

// SearchResponse is the response structure for search
type SearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string          `json:"_id"`
			Score  float64         `json:"_score"`
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// GetResponse is the response structure for get document
type GetResponse struct {
	Found  bool            `json:"found"`
	ID     string          `json:"_id"`
	Source json.RawMessage `json:"_source"`
}

// StatsResponse is the response structure for aggregations
type StatsResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
	} `json:"hits"`
	Aggregations struct {
		TotalProducts struct {
			Value int64 `json:"value"`
		} `json:"total_products"`
		TotalShops struct {
			Value int `json:"value"`
		} `json:"total_shops"`
	} `json:"aggregations"`
}
