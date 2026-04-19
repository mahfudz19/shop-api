// Package main adalah titik masuk utama (entry point) untuk aplikasi.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/username/shop-api/internal/config"

	"github.com/username/shop-api/internal/middleware"

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
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Printf("Warning: Gagal set trusted proxies: %v", err)
	}

	// 4. WIRING CLEAN ARCHITECTURE

	// ========== PRODUCT WIRING (Rute Publik) ==========
	// Endpoint seperti GET /products bisa diakses siapa saja tanpa login
	productRepo := productRepo.NewMongoProductRepository(db)
	productUsecase := productUsecase.NewProductUseCase(productRepo)
	productHttp.NewProductHandler(r, productUsecase)

	// ========== USER WIRING (Rute Publik) ==========
	// Endpoint /auth/login dan /auth/register harus publik
	userRepository := userRepo.NewMongoUserRepository(db)
	userUseCase := userUsecase.NewUserUseCase(userRepository)
	userHttp.NewUserHandler(r, userUseCase)

	// ========== CATEGORY WIRING (Rute Publik) ==========
	// Endpoint seperti GET /categories bisa diakses siapa saja tanpa login
	catRepo := categoryRepo.NewMongoCategoryRepository(db)
	catUC := categoryUseCase.NewCategoryUseCase(catRepo)
	categoryHttp.NewCategoryHandler(r, catUC)

	// ========== PROMOTION WIRING (Rute Publik) ==========
	// Endpoint seperti GET /promotions bisa diakses siapa saja tanpa login
	promoRepo := promotionRepo.NewMongoPromotionRepository(db)
	promoUC := promotionUseCase.NewPromotionUseCase(promoRepo)
	promotionHttp.NewPromotionHandler(r, promoUC)

	// ========== ARTICLE WIRING (Rute Publik) ==========
	// Endpoint seperti GET /articles bisa diakses siapa saja tanpa login
	articleRepository := articleRepo.NewMongoArticleRepository(db)
	articleUseCase := articleUseCase.NewArticleUseCase(articleRepository)
	articleHttp.NewArticleHandler(r, articleUseCase)

	// ========== MASTER PRODUCT WIRING (Rute Publik) ==========
	// Endpoint seperti GET /master-products/:id bisa diakses siapa saja tanpa login
	masterProductRepository := masterProductRepo.NewMongoMasterProductRepository(db)
	masterProductUseCase := masterProductUseCase.NewMasterProductUseCase(masterProductRepository)
	masterProductHttp.NewMasterProductHandler(r, masterProductUseCase)

	// ========== 5. RUTE TERPROTEKSI DENGAN MIDDLEWARE (BARU!) ==========
	// Buat grup rute baru yang diawali dengan "/admin"
	adminRoutes := r.Group("/admin")

	// Pasang satpam (middleware) hanya untuk grup rute ini
	adminRoutes.Use(middleware.AuthMiddleware())

	// Contoh endpoint yang sudah dilindungi satpam
	// Rute aslinya menjadi: GET /admin/dashboard
	adminRoutes.GET("/dashboard", func(c *gin.Context) {
		// Kita bisa mengambil ID pengguna dari context yang sudah disisipkan oleh middleware
		userID, _ := c.Get("user_id")
		userEmail, _ := c.Get("user_email")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Selamat datang di area admin!",
			"data": gin.H{
				"id":    userID,
				"email": userEmail,
			},
		})
	})

	// 6. Run Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server: http://localhost:%s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
