// Package service menyediakan service layer untuk operasi bisnis yang kompleks.
package service

import (
	"context"
	"fmt"
	"log"

	"github.com/username/shop-api/internal/domain"
)

// ElasticsearchProductRepository extends ProductRepository with ES-specific methods.
type ElasticsearchProductRepository interface {
	domain.ProductRepository
	IndexProduct(ctx context.Context, product domain.Product) error
	DeleteProduct(ctx context.Context, id string) error
}

// ElasticsearchSyncService handles synchronization between MongoDB and Elasticsearch.
type ElasticsearchSyncService struct {
	mongoRepo domain.ProductRepository       // Source of truth
	esRepo    ElasticsearchProductRepository // Destination (ES-specific interface)
}

// NewElasticsearchSyncService creates a new Elasticsearch sync service.
// mongoRepo: MongoDB repository (source of truth)
// esRepo: Elasticsearch repository (destination for search)
func NewElasticsearchSyncService(
	mongoRepo domain.ProductRepository,
	esRepo ElasticsearchProductRepository,
) *ElasticsearchSyncService {
	return &ElasticsearchSyncService{
		mongoRepo: mongoRepo,
		esRepo:    esRepo,
	}
}

// SyncAll syncs all products from MongoDB to Elasticsearch.
// Returns the number of successfully synced products.
// This method is designed for bulk/one-time sync operations.
func (s *ElasticsearchSyncService) SyncAll(ctx context.Context) (int, error) {
	log.Println("🔄 Fetching products from MongoDB...")

	// Fetch all products from MongoDB with large limit for bulk sync
	products, total, err := s.mongoRepo.FetchWithFilter(ctx, domain.ProductFilter{
		BaseQuery: domain.BaseQuery{
			Page:  1,
			Limit: 10000, // Large limit for bulk sync
		},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch products from MongoDB: %w", err)
	}

	if total == 0 {
		log.Println("⚠️  No products found in MongoDB")
		return 0, nil
	}

	log.Printf("📦 Found %d products in MongoDB, starting sync...", total)

	// Index each product to Elasticsearch
	synced := 0
	failed := 0
	for _, product := range products {
		if err := s.esRepo.IndexProduct(ctx, product); err != nil {
			log.Printf("⚠️  Failed to index product %s: %v", product.ID.Hex(), err)
			failed++
			continue
		}
		synced++
	}

	log.Printf("✅ Sync completed: %d/%d products synced (%d failed)", synced, total, failed)
	return synced, nil
}

// SyncSingle syncs a single product from MongoDB to Elasticsearch by ID.
// This method is designed for real-time sync (dual write) scenarios.
func (s *ElasticsearchSyncService) SyncSingle(ctx context.Context, id string) error {
	product, err := s.mongoRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch product %s from MongoDB: %w", id, err)
	}

	if err := s.esRepo.IndexProduct(ctx, product); err != nil {
		return fmt.Errorf("failed to index product %s to Elasticsearch: %w", id, err)
	}

	log.Printf("✅ Synced product %s to Elasticsearch", id)
	return nil
}

// SyncBatch syncs multiple products from MongoDB to Elasticsearch by IDs.
// Failed syncs are logged but don't stop the batch process.
func (s *ElasticsearchSyncService) SyncBatch(ctx context.Context, ids []string) error {
	log.Printf("🔄 Batch syncing %d products...", len(ids))

	success := 0
	failed := 0
	for _, id := range ids {
		if err := s.SyncSingle(ctx, id); err != nil {
			log.Printf("⚠️  Failed to sync product %s: %v", id, err)
			failed++
			continue
		}
		success++
	}

	log.Printf("✅ Batch sync completed: %d succeeded, %d failed", success, failed)
	return nil
}

// DeleteFromElasticsearch removes a product from Elasticsearch by ID.
// This method is designed for dual write delete scenarios.
func (s *ElasticsearchSyncService) DeleteFromElasticsearch(ctx context.Context, id string) error {
	if err := s.esRepo.DeleteProduct(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product %s from Elasticsearch: %w", id, err)
	}

	log.Printf("✅ Deleted product %s from Elasticsearch", id)
	return nil
}
