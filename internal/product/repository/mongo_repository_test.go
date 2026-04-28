package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/product/repository"
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

func TestMongoProductRepository_Integration(t *testing.T) {
	repo := repository.NewMongoProductRepository(testDB)
	ctx := context.Background()

	// Variabel global untuk relasi
	masterProductID := bson.NewObjectID()
	var targetProductID string // ID produk utama yang akan kita intip detailnya

	t.Run("0. Setup Data - Suntik Ekosistem Produk", func(t *testing.T) {
		// 1. Suntik Master Product
		_, err := testDB.Collection("master_products").InsertOne(ctx, bson.M{
			"_id":   masterProductID,
			"name":  "MacBook Pro M2",
			"brand": "Apple",
		})
		require.NoError(t, err)

		// 2. Siapkan data Produk yang kompleks
		waktu := time.Now().UTC()
		productsData := []interface{}{
			// Produk 1: Target Utama kita
			domain.Product{
				ID:              bson.NewObjectID(),
				Name:            "MacBook Pro M2 Resmi iBox",
				MasterProductID: masterProductID,
				PriceRp:         25000000,
				DiscountPercent: 0,
				Location:        "Jakarta",
				Marketplace:     "Tokopedia",
				Rating:          4.9,
				CreatedAt:       waktu,
				// Kita set field ini secara manual untuk test GetStats
			},
			// Produk 2: Offer Terkait (Murah)
			domain.Product{
				ID:              bson.NewObjectID(),
				Name:            "MacBook Pro M2 Inter",
				MasterProductID: masterProductID,
				PriceRp:         22000000,
				DiscountPercent: 10,
				Location:        "Batam",
				Marketplace:     "Shopee",
				Rating:          4.5,
				CreatedAt:       waktu.Add(time.Hour),
			},
			// Produk 3: Offer Terkait (Diskon Besar)
			domain.Product{
				ID:              bson.NewObjectID(),
				Name:            "MacBook Pro M2 Second",
				MasterProductID: masterProductID,
				PriceRp:         18000000,
				DiscountPercent: 20,
				Location:        "Jakarta",
				Marketplace:     "Tokopedia",
				Rating:          4.8,
				IsAnomaly:       false,
				CreatedAt:       waktu.Add(2 * time.Hour),
			},
			// Produk 4: Produk Lain (Bukan Master yang sama), Diskon Terbesar
			domain.Product{
				ID:              bson.NewObjectID(),
				Name:            "Laptop Gaming Asus",
				MasterProductID: bson.NewObjectID(), // Beda Master
				PriceRp:         15000000,
				DiscountPercent: 50,
				Location:        "Surabaya",
				Marketplace:     "Tokopedia",
				Rating:          4.0,
				CreatedAt:       waktu.Add(3 * time.Hour),
			},
		}

		// Agar field "shop" terbaca oleh GetStats, kita suntik BSON mentah
		_, err = testDB.Collection("products").InsertMany(ctx, []interface{}{
			bson.M{"_id": productsData[0].(domain.Product).ID, "name": productsData[0].(domain.Product).Name, "master_product_id": masterProductID, "price_rp": 25000000, "discount_percent": 0, "location": "Jakarta", "marketplace": "Tokopedia", "rating": 4.9, "shop": "iBox Official", "is_anomaly": false},
			bson.M{"_id": productsData[1].(domain.Product).ID, "name": productsData[1].(domain.Product).Name, "master_product_id": masterProductID, "price_rp": 22000000, "discount_percent": 10, "location": "Batam", "marketplace": "Shopee", "rating": 4.5, "shop": "BatamGadget", "is_anomaly": false},
			bson.M{"_id": productsData[2].(domain.Product).ID, "name": productsData[2].(domain.Product).Name, "master_product_id": masterProductID, "price_rp": 18000000, "discount_percent": 20, "location": "Jakarta", "marketplace": "Tokopedia", "rating": 4.8, "shop": "iBox Official", "is_anomaly": false},
			bson.M{"_id": productsData[3].(domain.Product).ID, "name": productsData[3].(domain.Product).Name, "master_product_id": productsData[3].(domain.Product).MasterProductID, "price_rp": 15000000, "discount_percent": 50, "location": "Surabaya", "marketplace": "Tokopedia", "rating": 4.0, "shop": "Asus Center", "is_anomaly": false},
		})
		require.NoError(t, err)

		targetProductID = productsData[0].(domain.Product).ID.Hex()
	})

	t.Run("1. FetchWithFilter - Test Regex Name & Pagination", func(t *testing.T) {
		filter := domain.ProductFilter{
			BaseQuery: domain.BaseQuery{
				Search: "macbook",
				Page:   1,
				Limit:  2,
			},
		}

		products, total, err := repo.FetchWithFilter(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, int64(3), total, "Total produk macbook ada 3, meskipun limitnya 2")
		assert.Len(t, products, 2, "Array yang direturn harus berjumlah sesuai limit (2)")
	})

	t.Run("2. FetchWithFilter - Test Multi-Filter Kompleks (Harga & Rating)", func(t *testing.T) {
		filter := domain.ProductFilter{
			BaseQuery: domain.BaseQuery{
				SortBy:    "price_rp",
				SortOrder: "desc",
			},
			MinPrice: 20000000,
			MaxPrice: 30000000,
			Rating:   4.8,
		}

		products, total, err := repo.FetchWithFilter(ctx, filter)

		require.NoError(t, err)
		// Yang lolos harga & rating:
		// 1. MacBook iBox (25 Juta, Rating 4.9) -> Lolos
		// 2. MacBook Inter (22 Juta, Rating 4.5) -> Gagal (Rating kurang)
		// 3. MacBook Second (18 Juta, Rating 4.8) -> Gagal (Harga kurang)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "MacBook Pro M2 Resmi iBox", products[0].Name)
	})

	t.Run("3. GetDeals - Test Filter Diskon & Sorting Tertinggi", func(t *testing.T) {
		deals, err := repo.GetDeals(ctx, 10)

		require.NoError(t, err)
		assert.Len(t, deals, 3, "Harusnya hanya 3 produk yang punya diskon > 0")

		// Produk diskon 50% harus di posisi ke-0
		assert.Equal(t, "Laptop Gaming Asus", deals[0].Name)
		assert.Equal(t, 50, deals[0].DiscountPercent)
		// Produk diskon 20% di posisi ke-1
		assert.Equal(t, 20, deals[1].DiscountPercent)
	})

	t.Run("4. GetStats - Test Aggregation (Unique Shops)", func(t *testing.T) {
		stats, err := repo.GetStats(ctx)

		require.NoError(t, err)
		assert.Equal(t, int64(4), stats.TotalProducts, "Total ada 4 produk")
		// Shop: "iBox Official", "BatamGadget", "iBox Official" (Duplikat), "Asus Center"
		assert.Equal(t, 3, stats.TotalShops, "Harus hanya ada 3 toko unik")
	})

	t.Run("5. GetByID - Test Cross Collection & Relasi", func(t *testing.T) {
		detail, err := repo.GetByID(ctx, targetProductID)

		require.NoError(t, err)

		// Validasi Produk Utama
		assert.Equal(t, "MacBook Pro M2 Resmi iBox", detail.Name)
	})
}
