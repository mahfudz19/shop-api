// Package main adalah titik masuk utama (entry point) untuk aplikasi.
package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/username/shop-api/internal/config"
	productHttp "github.com/username/shop-api/internal/product/delivery/http"
	"github.com/username/shop-api/internal/product/repository"
	"github.com/username/shop-api/internal/product/usecase"

	userHttp "github.com/username/shop-api/internal/user/delivery/http"
	userRepo "github.com/username/shop-api/internal/user/repository"
	userUsecase "github.com/username/shop-api/internal/user/usecase"
)

func main() {
	// 1. Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file tidak ditemukan")
	}

	// 2. CONNECT DATABASE (di sini!)
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

	// ========== PRODUCT WIRING (sudah ada) ==========
	productRepo := repository.NewMongoProductRepository(db)
	productUsecase := usecase.NewProductUseCase(productRepo)
	productHttp.NewProductHandler(r, productUsecase)

	// ========== USER WIRING (BARU!) ==========
	userRepository := userRepo.NewMongoUserRepository(db)
	userUseCase := userUsecase.NewUserUseCase(userRepository)
	userHttp.NewUserHandler(r, userUseCase)

	// 5. Run Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server: http://localhost:%s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
