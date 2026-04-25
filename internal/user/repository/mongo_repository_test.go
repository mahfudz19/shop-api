package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/testutil"
	"github.com/username/shop-api/internal/user/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Variabel global untuk menyimpan koneksi test DB
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

func TestMongoUserRepository_Integration(t *testing.T) {
	// Setup Repository dengan database test
	repo := repository.NewMongoUserRepository(testDB)
	ctx := context.Background()

	// Kita buat ID dan Email di awal agar bisa dipakai berurutan di sub-test
	userID := bson.NewObjectID()
	originalEmail := "Mahfudz.Engineer@Example.com" // Sengaja pakai huruf besar untuk test Collation

	testUser := domain.User{
		ID:        userID,
		Email:     originalEmail,
		Password:  "hashedpassword123",
		Name:      "Mahfudz",
		Role:      domain.RoleAdmin,
		Status:    domain.StatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	t.Run("1. Create - Sukses Insert User Baru", func(t *testing.T) {
		err := repo.Create(ctx, testUser)

		// Harusnya tidak ada error dari MongoDB
		require.NoError(t, err)
	})

	t.Run("2. GetByID - Sukses Ambil Data", func(t *testing.T) {
		// Gunakan Hex() karena fungsi GetByID menerima tipe string
		result, err := repo.GetByID(ctx, userID.Hex())

		require.NoError(t, err)
		assert.Equal(t, testUser.Name, result.Name)
		assert.Equal(t, testUser.Email, result.Email)
	})

	t.Run("3. GetByID - Gagal Format Hex ID Salah", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "id-bukan-hex-valid")

		assert.Error(t, err)
		// ✅ Perbarui pesan error agar sesuai dengan mongo-driver/v2
		assert.Contains(t, err.Error(), "not a valid ObjectID")
	})

	t.Run("4. GetByEmail - Sukses dengan Case-Insensitive", func(t *testing.T) {
		// Menguji kehebatan Collation MongoDB: Kita cari dengan huruf kecil semua!
		searchEmail := "mahfudz.engineer@example.com"

		result, err := repo.GetByEmail(ctx, searchEmail)

		require.NoError(t, err)
		// Membuktikan bahwa DB mengembalikan data yang benar meskipun input pencariannya beda kapital
		assert.Equal(t, originalEmail, result.Email)
	})

	t.Run("5. GetByEmail - Gagal User Tidak Ditemukan", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "hantu@tidakada.com")

		assert.Error(t, err)
		assert.Equal(t, mongo.ErrNoDocuments, err)
	})

	t.Run("6. EmailExists - Validasi True dan False", func(t *testing.T) {
		// Cek email yang ada (juga uji case-insensitive)
		exists, err := repo.EmailExists(ctx, "MAHFUDZ.engineer@example.com")
		require.NoError(t, err)
		assert.True(t, exists)

		// Cek email yang TIDAK ada
		exists, err = repo.EmailExists(ctx, "belum_daftar@example.com")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
