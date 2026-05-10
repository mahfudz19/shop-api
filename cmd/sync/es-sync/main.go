// Package main adalah entry point untuk script sync MongoDB → Elasticsearch.
// Script ini digunakan untuk one-time sync semua data produk.
package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/username/shop-api/internal/config"
	"github.com/username/shop-api/internal/product/repository"
	"github.com/username/shop-api/internal/service"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found, using environment variables")
	}

	ctx := context.Background()

	// ==========================================
	// 1. CONNECT TO MONGODB
	// ==========================================
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("❌ MONGODB_URI is required")
	}

	log.Println("🔌 Connecting to MongoDB...")
	mongoClient := config.ConnectMongoDB(mongoURI)
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("⚠️  Warning: Failed to disconnect MongoDB: %v", err)
		}
	}()

	// Get database name
	dbName := os.Getenv("MONGODB_NAME")
	if dbName == "" {
		dbName = "shop_db"
	}
	mongoDB := mongoClient.Database(dbName)

	log.Printf("✅ Connected to MongoDB: %s", dbName)

	// ==========================================
	// 2. CONNECT TO ELASTICSEARCH
	// ==========================================
	if os.Getenv("ELASTICSEARCH_ENABLED") != "true" {
		log.Fatal("❌ ELASTICSEARCH_ENABLED is not 'true'. Set ELASTICSEARCH_ENABLED=true in .env")
	}

	log.Println("🔌 Connecting to Elasticsearch...")
	esConfig := config.LoadElasticsearchConfig()
	esClient, err := config.NewElasticsearchClient(esConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Elasticsearch: %v", err)
	}

	log.Printf("✅ Connected to Elasticsearch: %s", esConfig.URL)

	// ==========================================
	// 3. INITIALIZE REPOSITORIES
	// ==========================================
	mongoRepo := repository.NewMongoProductRepository(mongoDB)

	// Create ES repository - returns concrete type with ES-specific methods
	esRepo := repository.NewElasticsearchProductRepository(esClient)

	// ==========================================
	// 4. CREATE SYNC SERVICE
	// ==========================================
	syncSvc := service.NewElasticsearchSyncService(mongoRepo, esRepo)

	// ==========================================
	// 5. RUN SYNC
	// ==========================================
	log.Println("🔄 Starting sync MongoDB → Elasticsearch...")
	log.Println("⏳ This may take a while depending on the number of products...")

	synced, err := syncSvc.SyncAll(ctx)
	if err != nil {
		log.Fatalf("❌ Sync failed: %v", err)
	}

	// ==========================================
	// 6. REPORT
	// ==========================================
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("✅ SYNC COMPLETED SUCCESSFULLY!\n")
	log.Printf("   📦 Total synced: %d products\n", synced)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("")
	log.Println("💡 Next steps:")
	log.Println("   1. Test the API: curl http://localhost:8080/products")
	log.Println("   2. Check Elasticsearch: make es-indices")
	log.Println("")
}
