// Package testutil menyediakan fungsi untuk mengelola koneksi ke MongoDB
package testutil

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/username/shop-api/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SetupTestDB menginisialisasi koneksi DB untuk testing dan fungsi pembersihnya.
func SetupTestDB() (*mongo.Database, func()) {
	uri := os.Getenv("MONGODB_TEST_URI")
	dbName := os.Getenv("MONGODB_TEST_NAME")

	// Sesuai logika Anda: Jika tidak ada env, batalkan test (Fail Fast)
	if uri == "" || dbName == "" {
		log.Fatal("❌ ERROR: MONGODB_TEST_URI dan MONGODB_TEST_NAME harus diisi untuk menjalankan test!")
	}

	// Gunakan fungsi config yang sudah ada agar Clean Architecture tetap terjaga
	client := config.ConnectMongoDB(uri)
	db := client.Database(dbName)

	// Fungsi Cleanup: Ini akan dipanggil di akhir test untuk membersihkan "sampah" data
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Println("🧹 Membersihkan database test (Drop)...")
		if err := db.Drop(ctx); err != nil {
			log.Printf("⚠️ Gagal drop test database: %v", err)
		}

		_ = client.Disconnect(ctx)
		log.Println("🔌 Koneksi database test ditutup.")
	}

	return db, cleanup
}
