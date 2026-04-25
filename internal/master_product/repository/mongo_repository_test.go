package repository_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/username/shop-api/internal/master_product/repository"
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

func TestMongoMasterProductRepository_Integration(t *testing.T) {
	repo := repository.NewMongoMasterProductRepository(testDB)
	ctx := context.Background()

	// Setup: ID Master Product yang akan kita jadikan patokan
	masterProductID := bson.NewObjectID()

	t.Run("0. Setup Data - Suntik Data BSON Mentah", func(t *testing.T) {
		// Karena repo tidak punya fungsi Create, kita suntik manual sebagai prasyarat test

		// 1. Suntik 1 Master Product
		_, err := testDB.Collection("master_products").InsertOne(ctx, bson.M{
			"_id":   masterProductID,
			"name":  "iPhone 15 Pro Max",
			"brand": "Apple",
		})
		require.NoError(t, err)

		// 2. Suntik Penawaran (Offers) ke koleksi "products"
		offersData := []interface{}{
			// Offer A: Valid, Harga Tengah
			bson.M{
				"master_product_id": masterProductID,
				"name":              "iPhone 15 Pro Max Garansi iBox",
				"price_rp":          25000000,
				"is_anomaly":        false,
				"match_confidence":  0.95, // Lolos filter (>= 0.8)
			},
			// Offer B: Valid, Harga Termurah (Harusnya muncul pertama)
			bson.M{
				"master_product_id": masterProductID,
				"name":              "iPhone 15 Pro Max Inter",
				"price_rp":          23000000,
				"is_anomaly":        false,
				"match_confidence":  0.85, // Lolos filter
			},
			// Offer C: INVALID (Anomaly)
			bson.M{
				"master_product_id": masterProductID,
				"name":              "Casing iPhone 15",
				"price_rp":          150000,
				"is_anomaly":        true, // Kena filter anomali
				"match_confidence":  0.90,
			},
			// Offer D: INVALID (Confidence Rendah)
			bson.M{
				"master_product_id": masterProductID,
				"name":              "Kredit iPhone 15",
				"price_rp":          2000000,
				"is_anomaly":        false,
				"match_confidence":  0.50, // Kena filter confidence (< 0.8)
			},
			// Offer E: INVALID (Beda Master Product)
			bson.M{
				"master_product_id": bson.NewObjectID(), // ID lain
				"name":              "Samsung S24 Ultra",
				"price_rp":          20000000,
				"is_anomaly":        false,
				"match_confidence":  0.99,
			},
		}

		_, err = testDB.Collection("products").InsertMany(ctx, offersData)
		require.NoError(t, err)
	})

	t.Run("1. GetByID - Sukses Ambil Data Master", func(t *testing.T) {
		master, err := repo.GetByID(ctx, masterProductID.Hex())

		require.NoError(t, err)
		assert.Equal(t, "iPhone 15 Pro Max", master.Name)
		assert.Equal(t, "Apple", master.Brand)
	})

	t.Run("2. GetByID - Gagal Format Hex Salah", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "bukan-hex-valid")

		assert.Error(t, err)
		assert.Equal(t, "invalid master product ID format", err.Error())
	})

	t.Run("3. GetByID - Gagal Master Product Tidak Ditemukan", func(t *testing.T) {
		randomID := bson.NewObjectID().Hex()
		_, err := repo.GetByID(ctx, randomID)

		assert.Error(t, err)
		assert.Equal(t, "master product not found", err.Error())
	})

	t.Run("4. GetOffersByMasterID - Filter Anomali & Sorting Harga Bekerja Sempurna", func(t *testing.T) {
		offers, err := repo.GetOffersByMasterID(ctx, masterProductID)

		require.NoError(t, err)

		// EKSPEKTASI FILTER: Dari 5 data yang disuntikkan, hanya 2 yang sah untuk Master Product ini.
		// (1 Anomali, 1 Confidence < 0.8, dan 1 milik HP Samsung dibuang oleh MongoDB).
		require.Len(t, offers, 2, "Harusnya hanya mengembalikan 2 penawaran yang valid")

		// EKSPEKTASI SORTING: Offer B (Rp 23.000.000) harus berada di index 0 karena paling murah.
		assert.Equal(t, int64(23000000), offers[0].PriceRp, "Harga termurah harus berada di urutan pertama")
		assert.Equal(t, int64(25000000), offers[1].PriceRp)
	})
}
