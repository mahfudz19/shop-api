package config

import (
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

// ElasticsearchConfig holds Elasticsearch configuration
type ElasticsearchConfig struct {
	URL    string
	APIKey string
	Index  string
}

// LoadElasticsearchConfig loads configuration from environment variables
func LoadElasticsearchConfig() ElasticsearchConfig {
	return ElasticsearchConfig{
		URL:    getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		APIKey: getEnv("ELASTICSEARCH_API_KEY", ""),
		Index:  getEnv("ELASTICSEARCH_INDEX", "products"),
	}
}

// NewElasticsearchClient creates a new Elasticsearch client
func NewElasticsearchClient(cfg ElasticsearchConfig) (*elasticsearch.Client, error) {
	config := elasticsearch.Config{
		Addresses: []string{cfg.URL},
		APIKey:    cfg.APIKey,
	}

	client, err := elasticsearch.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("error connecting to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch connection error: %s", res.String())
	}

	log.Printf("✅ Connected to Elasticsearch: %s", res.String())
	return client, nil
}

// getEnv returns environment variable value or default if not set
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
