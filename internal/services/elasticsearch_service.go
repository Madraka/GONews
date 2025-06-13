package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"news/internal/json"
	"news/internal/models"
	"news/internal/tracing"

	"go.opentelemetry.io/otel/attribute"
)

// Global ElasticSearch service instance
var elasticsearchServiceInstance *ElasticSearchService

// ElasticSearchService handles vector search operations with ElasticSearch
type ElasticSearchService struct {
	baseURL    string
	username   string
	password   string
	indexName  string
	httpClient *http.Client
}

// NewElasticSearchService creates a new ElasticSearch service instance
func NewElasticSearchService() *ElasticSearchService {
	return &ElasticSearchService{
		baseURL:   getESEnvOrDefault("ELASTICSEARCH_URL", "http://localhost:9200"),
		username:  os.Getenv("ELASTICSEARCH_USERNAME"),
		password:  os.Getenv("ELASTICSEARCH_PASSWORD"),
		indexName: getESEnvOrDefault("ELASTICSEARCH_INDEX", "news-articles"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetElasticSearchService returns the global ElasticSearch service instance
func GetElasticSearchService() *ElasticSearchService {
	if elasticsearchServiceInstance == nil {
		elasticsearchServiceInstance = NewElasticSearchService()
	}
	return elasticsearchServiceInstance
}

// InitializeIndex creates the ElasticSearch index with proper mapping for dense vectors
func (es *ElasticSearchService) InitializeIndex(ctx context.Context) error {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "elasticsearch.initialize_index",
		attribute.String("index_name", es.indexName))
	defer span.End()

	// Check if index exists
	exists, err := es.indexExists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if exists {
		span.SetAttributes(attribute.Bool("index_already_exists", true))
		return nil
	}

	// Create index with mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"title": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"summary": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"content": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"published_at": map[string]interface{}{
					"type": "date",
				},
				"lang": map[string]interface{}{
					"type": "keyword",
				},
				"region": map[string]interface{}{
					"type": "keyword",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
				"embedding": map[string]interface{}{
					"type":       "dense_vector",
					"dims":       1536, // OpenAI text-embedding-3-small dimension
					"index":      true,
					"similarity": "cosine",
				},
			},
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	}

	return es.makeRequest(ctx, "PUT", fmt.Sprintf("/%s", es.indexName), mapping, nil)
}

// IndexDocument indexes a document with its embedding vector
func (es *ElasticSearchService) IndexDocument(ctx context.Context, doc *models.SearchDocument) error {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "elasticsearch.index_document",
		attribute.String("document_id", doc.ID),
		attribute.String("document_title", doc.Title))
	defer span.End()

	url := fmt.Sprintf("/%s/_doc/%s", es.indexName, doc.ID)
	return es.makeRequest(ctx, "PUT", url, doc, nil)
}

// VectorSearch performs semantic search using vector similarity
func (es *ElasticSearchService) VectorSearch(ctx context.Context, req *models.SemanticSearchRequest) (*models.SemanticSearchResponse, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "elasticsearch.vector_search",
		attribute.String("query", req.Query),
		attribute.Int("limit", req.Limit))
	defer span.End()

	if len(req.Embedding) == 0 {
		return nil, fmt.Errorf("embedding vector is required for semantic search")
	}

	// Build the search query
	query := map[string]interface{}{
		"size": req.Limit,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"script_score": map[string]interface{}{
							"query": map[string]interface{}{
								"match_all": map[string]interface{}{},
							},
							"script": map[string]interface{}{
								"source": "cosineSimilarity(params.query_vector, 'embedding') + 1.0",
								"params": map[string]interface{}{
									"query_vector": req.Embedding,
								},
							},
						},
					},
				},
				"filter": es.buildFilters(req),
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"_score": map[string]interface{}{
					"order": "desc",
				},
			},
			map[string]interface{}{
				"published_at": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"_source": []string{"id", "title", "summary", "published_at", "lang", "region", "category"},
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string                 `json:"_id"`
				Score  float64                `json:"_score"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	err := es.makeRequest(ctx, "POST", fmt.Sprintf("/%s/_search", es.indexName), query, &result)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Convert results
	response := &models.SemanticSearchResponse{
		Query:   req.Query, // Fix: Set the query field from the request
		Results: make([]models.SemanticSearchResult, len(result.Hits.Hits)),
		Total:   result.Hits.Total.Value,
	}

	for i, hit := range result.Hits.Hits {
		// Parse published_at from string to time.Time
		publishedAt, _ := time.Parse(time.RFC3339, getString(hit.Source, "published_at"))

		response.Results[i] = models.SemanticSearchResult{
			ID:          hit.ID,
			Title:       getString(hit.Source, "title"),
			Summary:     getString(hit.Source, "summary"),
			PublishedAt: publishedAt,
			Score:       hit.Score,
			Lang:        getString(hit.Source, "lang"),
			Region:      getString(hit.Source, "region"),
		}
	}

	span.SetAttributes(
		attribute.Int("results_count", len(response.Results)),
		attribute.Int("total_hits", response.Total))

	return response, nil
}

// FallbackSearch performs traditional text-based search when vector search fails
func (es *ElasticSearchService) FallbackSearch(ctx context.Context, req *models.SemanticSearchRequest) (*models.SemanticSearchResponse, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "elasticsearch.fallback_search",
		attribute.String("query", req.Query),
		attribute.Int("limit", req.Limit))
	defer span.End()

	// Build multi-match query for fallback
	query := map[string]interface{}{
		"size": req.Limit,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query":  req.Query,
							"fields": []string{"title^3", "summary^2", "content"},
							"type":   "best_fields",
						},
					},
				},
				"filter": es.buildFilters(req),
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"_score": map[string]interface{}{
					"order": "desc",
				},
			},
			map[string]interface{}{
				"published_at": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"_source": []string{"id", "title", "summary", "published_at", "lang", "region", "category"},
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string                 `json:"_id"`
				Score  float64                `json:"_score"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	err := es.makeRequest(ctx, "POST", fmt.Sprintf("/%s/_search", es.indexName), query, &result)
	if err != nil {
		return nil, fmt.Errorf("fallback search failed: %w", err)
	}

	// Convert results
	response := &models.SemanticSearchResponse{
		Query:   req.Query, // Fix: Set the query field from the request
		Results: make([]models.SemanticSearchResult, len(result.Hits.Hits)),
		Total:   result.Hits.Total.Value,
	}

	for i, hit := range result.Hits.Hits {
		// Parse published_at from string to time.Time
		publishedAt, _ := time.Parse(time.RFC3339, getString(hit.Source, "published_at"))

		response.Results[i] = models.SemanticSearchResult{
			ID:          hit.ID,
			Title:       getString(hit.Source, "title"),
			Summary:     getString(hit.Source, "summary"),
			PublishedAt: publishedAt,
			Score:       hit.Score,
			Lang:        getString(hit.Source, "lang"),
			Region:      getString(hit.Source, "region"),
		}
	}

	span.SetAttributes(
		attribute.Int("results_count", len(response.Results)),
		attribute.Int("total_hits", response.Total))

	return response, nil
}

// buildFilters creates ElasticSearch filters from the request
func (es *ElasticSearchService) buildFilters(req *models.SemanticSearchRequest) []interface{} {
	var filters []interface{}

	if req.Lang != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"lang": req.Lang,
			},
		})
	}

	if req.Region != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"region": req.Region,
			},
		})
	}

	if req.Category != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"category": req.Category,
			},
		})
	}

	return filters
}

// indexExists checks if the ElasticSearch index exists
func (es *ElasticSearchService) indexExists(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", es.baseURL+"/"+es.indexName, nil)
	if err != nil {
		return false, err
	}

	if es.username != "" && es.password != "" {
		req.SetBasicAuth(es.username, es.password)
	}

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close Elasticsearch response body: %v", err)
		}
	}()

	return resp.StatusCode == http.StatusOK, nil
}

// makeRequest makes an HTTP request to ElasticSearch
func (es *ElasticSearchService) makeRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, es.baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if es.username != "" && es.password != "" {
		req.SetBasicAuth(es.username, es.password)
	}

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close Elasticsearch request response body: %v", err)
		}
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("elasticsearch request failed with status %d", resp.StatusCode)
	}

	if result != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// Helper function to get environment variable with default
func getESEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
