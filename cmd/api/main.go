// Package main adalah titik masuk utama (entry point) untuk aplikasi.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/username/shop-api/internal/config"
	"github.com/username/shop-api/internal/service"

	productHttp "github.com/username/shop-api/internal/product/delivery/http"
	"github.com/username/shop-api/internal/product/repository"
	productUsecase "github.com/username/shop-api/internal/product/usecase"

	masterProductHttp "github.com/username/shop-api/internal/master_product/delivery/http"
	masterProductRepo "github.com/username/shop-api/internal/master_product/repository"
	masterProductUseCase "github.com/username/shop-api/internal/master_product/usecase"

	userHttp "github.com/username/shop-api/internal/user/delivery/http"
	userRepo "github.com/username/shop-api/internal/user/repository"
	userUsecase "github.com/username/shop-api/internal/user/usecase"

	categoryHttp "github.com/username/shop-api/internal/category/delivery/http"
	categoryRepo "github.com/username/shop-api/internal/category/repository"
	categoryUseCase "github.com/username/shop-api/internal/category/usecase"

	promotionHttp "github.com/username/shop-api/internal/promotion/delivery/http"
	promotionRepo "github.com/username/shop-api/internal/promotion/repository"
	promotionUseCase "github.com/username/shop-api/internal/promotion/usecase"

	articleHttp "github.com/username/shop-api/internal/article/delivery/http"
	articleRepo "github.com/username/shop-api/internal/article/repository"
	articleUseCase "github.com/username/shop-api/internal/article/usecase"

	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/middleware"
)

func main() {
	// 1. Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file tidak ditemukan")
	}

	// 2. CONNECT DATABASE
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("❌ MONGODB_URI tidak ditemukan di .env")
	}

	client := config.ConnectMongoDB(mongoURI)

	// Cleanup saat program exit
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Println("Warning: Gagal disconnect:", err)
		}
	}()

	// Ambil database
	dbName := os.Getenv("MONGODB_NAME")
	if dbName == "" {
		dbName = "shop_db"
	}
	db := client.Database(dbName)

	log.Printf("📦 Database: %s", dbName)
	// Initialize Elasticsearch
	var esClient *elasticsearch.Client
	var err error

	if os.Getenv("ELASTICSEARCH_ENABLED") == "true" {
		esCfg := config.LoadElasticsearchConfig()
		esClient, err = config.NewElasticsearchClient(esCfg)
		if err != nil {
			log.Printf("⚠️ Warning: Elasticsearch not available, using MongoDB only: %v", err)
		}
	}

	// 3. Init Repositories
	// MongoDB Product Repository (source of truth)
	mongoProductRepo := repository.NewMongoProductRepository(db)

	// Elasticsearch Product Repository (for search)
	var esProductRepo *repository.ElasticsearchProductRepository
	if esClient != nil {
		esProductRepo = repository.NewElasticsearchProductRepository(esClient)
		log.Println("✅ Elasticsearch repository initialized")
	}

	// Create Sync Service (if ES is available)
	var syncSvc *service.ElasticsearchSyncService
	if esProductRepo != nil {
		syncSvc = service.NewElasticsearchSyncService(mongoProductRepo, esProductRepo)
		log.Println("✅ Elasticsearch Sync Service initialized")
	}

	// Product Repository - use ES if available, fallback to MongoDB
	var productRepository domain.ProductRepository
	if esProductRepo != nil {
		productRepository = esProductRepo
		log.Println("✅ Using Elasticsearch repository for queries")
	} else {
		productRepository = mongoProductRepo
		log.Println("✅ Using MongoDB repository for queries")
	}

	// Other Repositories
	userRepository := userRepo.NewMongoUserRepository(db)
	masterProductRepository := masterProductRepo.NewMongoMasterProductRepository(db)
	catRepo := categoryRepo.NewMongoCategoryRepository(db)
	promoRepo := promotionRepo.NewMongoPromotionRepository(db)
	articleRepository := articleRepo.NewMongoArticleRepository(db)

	// 4. Init Gin
	r := gin.Default()
	r.Use(middleware.SecurityHeaders())
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 🛡️ SECURITY FIX: Matikan "Trust All Proxies" bawaan Gin
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Printf("⚠️ Warning: Gagal mengeset trusted proxies: %v\n", err)
	}

	// 4. Rate Limiter - HANYA aktif di production (bukan development)
	appEnv := os.Getenv("APP_ENV")
	if appEnv != "development" {
		r.Use(middleware.RateLimiter())
		log.Println("🛡️  Rate Limiter: ACTIVE (production mode)")
	} else {
		log.Println("⚠️  Rate Limiter: DISABLED (development mode)")
	}

	r.GET("/", func(c *gin.Context) {
		content := fmt.Sprintf("online for %s", os.Getenv("APP_ENV"))
		c.String(200, content)
	})

	// ==========================================
	// 5. PEMBUATAN GRUP RUTE (ROUTER GROUPS)
	// ==========================================

	// Grup Publik (Tanpa Middleware, siapa saja bisa akses)
	publicRoutes := r.Group("")

	// Grup Protected (Wajib Login)
	protectedRoutes := r.Group("")
	protectedRoutes.Use(middleware.AuthMiddleware())
	protectedRoutes.Use(middleware.CSRFProtection())

	// Grup Admin (Wajib Login + Wajib Role Admin)
	adminRoutes := r.Group("")
	adminRoutes.Use(middleware.AuthMiddleware())
	adminRoutes.Use(middleware.RequireRole(domain.RoleAdmin))
	adminRoutes.Use(middleware.CSRFProtection())

	// ==========================================
	// 6. WIRING USECASES & HANDLERS
	// ==========================================

	// User
	userUseCase := userUsecase.NewUserUseCase(userRepository)
	userHttp.NewUserHandler(publicRoutes, protectedRoutes, adminRoutes, userUseCase)

	// Product
	productUseCase := productUsecase.NewProductUseCase(productRepository)
	productHttp.NewProductHandler(publicRoutes, adminRoutes, productUseCase, syncSvc)

	// Master Product
	masterProductUseCase := masterProductUseCase.NewMasterProductUseCase(masterProductRepository)
	masterProductHttp.NewMasterProductHandler(publicRoutes, protectedRoutes, masterProductUseCase)

	// Category
	catUseCase := categoryUseCase.NewCategoryUseCase(catRepo)
	categoryHttp.NewCategoryHandler(publicRoutes, protectedRoutes, catUseCase)

	// Promotion
	promoUseCase := promotionUseCase.NewPromotionUseCase(promoRepo)
	promotionHttp.NewPromotionHandler(publicRoutes, protectedRoutes, promoUseCase)

	// Article
	articleUseCase := articleUseCase.NewArticleUseCase(articleRepository)
	articleHttp.NewArticleHandler(publicRoutes, protectedRoutes, articleUseCase)

	// ==========================================
	// 7. Run Server
	// ==========================================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server: http://localhost:%s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
