package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/promotion/repository"
	"github.com/username/shop-api/internal/testutil"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var testDB *mongo.Database

// TestMain bertindak sebagai konduktor untuk seluruh file test di package ini
func TestMain(m *testing.M) {
	// 1. Panggil reusable connection dari testutil
	db, cleanup := testutil.SetupTestDB()
	testDB = db

	// 2. Jalankan semua fungsi test
	code := m.Run()

	// 3. Eksekusi fungsi cleanup (Drop DB & Disconnect)
	cleanup()

	os.Exit(code)
}

func TestMongoPromotionRepository_Integration(t *testing.T) {
	repo := repository.NewMongoPromotionRepository(testDB)
	ctx := context.Background()

	// Variabel global dalam scope test ini untuk memindahkan ID antar skenario
	var createdPromoID string

	// Data uji: Kita sengaja membuat urutan yang acak untuk menguji fitur Sorting MongoDB
	promoA := domain.Promotion{
		Title:      "Promo Akhir Tahun",
		IsActive:   true,
		OrderIndex: 2, // Urutan kedua
		CreatedAt:  time.Now().UTC(),
	}
	promoB := domain.Promotion{
		Title:      "Promo Flash Sale",
		IsActive:   true,
		OrderIndex: 1, // Urutan pertama
		CreatedAt:  time.Now().UTC(),
	}
	promoInactive := domain.Promotion{
		Title:      "Promo Kedaluwarsa",
		IsActive:   false,
		OrderIndex: 3, // Urutan ketiga
		CreatedAt:  time.Now().UTC(),
	}

	t.Run("1. Create - Sukses Insert Beberapa Promosi", func(t *testing.T) {
		// Insert Promo A
		resA, err := repo.Create(ctx, promoA)
		require.NoError(t, err)
		createdPromoID = resA.ID.Hex() // Simpan ID untuk test Update & Delete nanti

		// Insert Promo B & Inactive
		_, err = repo.Create(ctx, promoB)
		require.NoError(t, err)
		_, err = repo.Create(ctx, promoInactive)
		require.NoError(t, err)
	})

	t.Run("2. GetAll - Test Sorting dan Filter (Hanya Aktif)", func(t *testing.T) {
		// Test ambil HANYA promo yang aktif
		promos, err := repo.GetAll(ctx, true)

		require.NoError(t, err)
		assert.Len(t, promos, 2, "Harusnya hanya return 2 promo yang IsActive = true")

		// Membuktikan fitur Sorting: Promo B (OrderIndex 1) harusnya di urutan index ke-0
		assert.Equal(t, promoB.Title, promos[0].Title, "Promo dengan OrderIndex terkecil harus muncul pertama")
		assert.Equal(t, promoA.Title, promos[1].Title)
	})

	t.Run("3. GetAll - Test Tanpa Filter (Semua Data)", func(t *testing.T) {
		// Test ambil SEMUA promo (termasuk yang tidak aktif)
		promos, err := repo.GetAll(ctx, false)

		require.NoError(t, err)
		assert.Len(t, promos, 3, "Harusnya mengembalikan semua 3 promo")
	})

	t.Run("4. GetByID - Validasi Data Spesifik", func(t *testing.T) {
		promo, err := repo.GetByID(ctx, createdPromoID)

		require.NoError(t, err)
		assert.Equal(t, promoA.Title, promo.Title)
		assert.Equal(t, true, promo.IsActive)
	})

	t.Run("5. GetByID - Gagal Format Hex Salah", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "format-id-palsu")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid id") // Berdasarkan custom error Anda "invalid id"
	})

	t.Run("6. Update - Sukses Mengubah Status dan Return Data Baru", func(t *testing.T) {
		// Kita akan me-nonaktifkan Promo A
		updatePayload := domain.Promotion{
			Title:      "Promo Akhir Tahun (SELESAI)",
			IsActive:   false, // Ubah status
			OrderIndex: 2,
		}

		updatedPromo, err := repo.Update(ctx, createdPromoID, updatePayload)

		require.NoError(t, err)
		// Membuktikan `options.After` bekerja dengan memverifikasi data yang direturn adalah yang terbaru
		assert.Equal(t, "Promo Akhir Tahun (SELESAI)", updatedPromo.Title)
		assert.Equal(t, false, updatedPromo.IsActive)
	})

	t.Run("7. Delete - Sukses Menghapus dan Verifikasi Hilang", func(t *testing.T) {
		// Hapus data
		err := repo.Delete(ctx, createdPromoID)
		require.NoError(t, err)

		// Coba GetByID lagi, harusnya error ErrNoDocuments
		_, err = repo.GetByID(ctx, createdPromoID)
		assert.Error(t, err)
		assert.Equal(t, mongo.ErrNoDocuments, err)
	})
}
