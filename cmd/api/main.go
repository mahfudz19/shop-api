// Package main adalah titik masuk utama (entry point) untuk aplikasi.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/username/shop-api/internal/config"

	productHttp "github.com/username/shop-api/internal/product/delivery/http"
	productRepo "github.com/username/shop-api/internal/product/repository"
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

	// 3. Init Gin
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		content := fmt.Sprintf("online for %s", os.Getenv("APP_ENV"))
		c.String(200, content)
	})

	// 4. Rate Limiter - HANYA aktif di production (bukan development)
	appEnv := os.Getenv("APP_ENV")
	if appEnv != "development" {
		r.Use(middleware.RateLimiter())
		log.Println("🛡️  Rate Limiter: ACTIVE (production mode)")
	} else {
		log.Println("⚠️  Rate Limiter: DISABLED (development mode)")
	}

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
	// 6. WIRING HANDLERS & USECASES
	// ==========================================

	// User
	userRepository := userRepo.NewMongoUserRepository(db)
	userUseCase := userUsecase.NewUserUseCase(userRepository)
	userHttp.NewUserHandler(publicRoutes, protectedRoutes, adminRoutes, userUseCase)

	// Product
	productRepository := productRepo.NewMongoProductRepository(db)
	productUsecase := productUsecase.NewProductUseCase(productRepository)
	productHttp.NewProductHandler(publicRoutes, adminRoutes, productUsecase)

	// Master Product
	masterProductRepository := masterProductRepo.NewMongoMasterProductRepository(db)
	masterProductUseCase := masterProductUseCase.NewMasterProductUseCase(masterProductRepository)
	masterProductHttp.NewMasterProductHandler(publicRoutes, protectedRoutes, masterProductUseCase)

	// Category
	catRepo := categoryRepo.NewMongoCategoryRepository(db)
	catUC := categoryUseCase.NewCategoryUseCase(catRepo)
	categoryHttp.NewCategoryHandler(publicRoutes, protectedRoutes, catUC)

	// Promotion
	promoRepo := promotionRepo.NewMongoPromotionRepository(db)
	promoUC := promotionUseCase.NewPromotionUseCase(promoRepo)
	promotionHttp.NewPromotionHandler(publicRoutes, protectedRoutes, promoUC)

	// Article
	articleRepository := articleRepo.NewMongoArticleRepository(db)
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
