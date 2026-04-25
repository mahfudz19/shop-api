package repository_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/username/shop-api/internal/category/repository"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/testutil"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var testDB *mongo.Database

func TestMain(m *testing.M) {
	db, cleanup := testutil.SetupTestDB()
	testDB = db
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func TestMongoCategoryRepository_Integration(t *testing.T) {
	repo := repository.NewMongoCategoryRepository(testDB)
	ctx := context.Background()

	var createdCategoryID string

	// Data uji dengan urutan (OrderIndex) yang dibalik
	catZebra := domain.Category{
		Name:       "Zebra",
		Slug:       "zebra",
		OrderIndex: 2, // Masuk pertama, tapi urutan 2
	}
	catAyam := domain.Category{
		Name:       "Ayam",
		Slug:       "ayam",
		OrderIndex: 1, // Masuk belakangan, tapi urutan 1
	}

	t.Run("1. Create - Sukses Insert Kategori", func(t *testing.T) {
		resZebra, err := repo.Create(ctx, catZebra)
		require.NoError(t, err)
		createdCategoryID = resZebra.ID.Hex() // Simpan ID Zebra

		_, err = repo.Create(ctx, catAyam)
		require.NoError(t, err)
	})

	t.Run("2. GetAll - Test Sorting Ascending (OrderIndex)", func(t *testing.T) {
		cats, err := repo.GetAll(ctx)

		require.NoError(t, err)
		assert.Len(t, cats, 2)

		// Membuktikan fitur sorting MongoDB bekerja: Kategori "Ayam" (Index 1) harus muncul di array [0]
		assert.Equal(t, catAyam.Name, cats[0].Name, "Kategori dengan OrderIndex terkecil harus di urutan pertama")
		assert.Equal(t, catZebra.Name, cats[1].Name)
	})

	t.Run("3. GetByID - Validasi Data Spesifik", func(t *testing.T) {
		cat, err := repo.GetByID(ctx, createdCategoryID) // Mencari Zebra

		require.NoError(t, err)
		assert.Equal(t, catZebra.Name, cat.Name)
	})

	t.Run("4. GetByID - Gagal Category Not Found", func(t *testing.T) {
		// Menggunakan ID hex valid tapi tidak ada di DB
		randomID := bson.NewObjectID().Hex()
		_, err := repo.GetByID(ctx, randomID)

		assert.Error(t, err)
		assert.Equal(t, "category not found", err.Error())
	})

	t.Run("5. Update - Sukses Mengubah Data", func(t *testing.T) {
		updatePayload := domain.Category{
			Name:       "Zebra Cross",
			Slug:       "zebra-cross",
			OrderIndex: 5,
		}

		updatedCat, err := repo.Update(ctx, createdCategoryID, updatePayload)

		require.NoError(t, err)
		assert.Equal(t, "Zebra Cross", updatedCat.Name)
	})

	t.Run("6. Delete - Sukses Menghapus Data", func(t *testing.T) {
		err := repo.Delete(ctx, createdCategoryID) // Menghapus Zebra Cross
		require.NoError(t, err)

		// Saat ini di DB hanya tersisa kategori "Ayam"
	})

	t.Run("7. Delete - Gagal Kategori Tidak Ditemukan", func(t *testing.T) {
		// Mencoba menghapus Zebra lagi
		err := repo.Delete(ctx, createdCategoryID)

		assert.Error(t, err)
		assert.Equal(t, "category not found", err.Error())
	})

	t.Run("8. SyncCategories - Fitur Cerdas Pemindahan Koleksi Silang", func(t *testing.T) {
		// SETUP: Kita tembak langsung BSON mentah ke koleksi "products"
		productsColl := testDB.Collection("products")
		_, err := productsColl.InsertMany(ctx, []interface{}{
			bson.M{"name": "Laptop", "category": "Elektronik"},
			bson.M{"name": "Baju", "category": "Fashion"},
			bson.M{"name": "Mouse", "category": "Elektronik"}, // Duplikat Elektronik
			bson.M{"name": "Ayam Goreng", "category": "Ayam"}, // Kategori Ayam sudah ada di DB
		})
		require.NoError(t, err)

		// EKSEKUSI: Panggil Usecase Sinkronisasi
		insertedCount, err := repo.SyncCategories(ctx)

		require.NoError(t, err)
		// EKSPEKTASI:
		// Distinct dari product = [Elektronik, Fashion, Ayam]
		// Existing di categories DB = [Ayam] (Karena Zebra sudah kita hapus di skenario 6)
		// Yang benar-benar baru = [Elektronik, Fashion] -> Total 2
		assert.Equal(t, int64(2), insertedCount)

		// Validasi akhir ke DB
		finalCats, _ := repo.GetAll(ctx)
		assert.Len(t, finalCats, 3) // Ayam, Elektronik, Fashion
	})
}
