package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/username/shop-api/internal/article/repository"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/testutil"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var testDB *mongo.Database

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

func TestMongoArticleRepository_Integration(t *testing.T) {
	repo := repository.NewMongoArticleRepository(testDB)
	ctx := context.Background()

	// Variabel global untuk menyimpan state antar skenario
	var createdArticleID string
	targetSlug := "5-tips-koding-golang"

	// Data uji: Kita manipulasi waktunya untuk menguji sorting Descending (-1)
	waktuSekarang := time.Now().UTC()
	waktuKemarin := waktuSekarang.Add(-24 * time.Hour)

	articleLama := domain.Article{
		Title:       "Artikel Kemarin",
		Slug:        "artikel-kemarin",
		Content:     "Isi kemarin",
		IsPublished: true,
		CreatedAt:   waktuKemarin, // Lebih lama
	}

	articleBaru := domain.Article{
		Title:       "5 Tips Koding Golang",
		Slug:        targetSlug,
		Content:     "Isi tips golang",
		IsPublished: true,
		CreatedAt:   waktuSekarang, // Lebih baru
	}

	articleDraft := domain.Article{
		Title:       "Rahasia Belum Rilis",
		Slug:        "rahasia-belum-rilis",
		Content:     "Masih diketik",
		IsPublished: false, // Draft
		CreatedAt:   waktuSekarang,
	}

	t.Run("1. Create - Sukses Insert Beberapa Artikel", func(t *testing.T) {
		// Insert Artikel Lama
		_, err := repo.Create(ctx, articleLama)
		require.NoError(t, err)

		// Insert Artikel Baru
		resBaru, err := repo.Create(ctx, articleBaru)
		require.NoError(t, err)
		createdArticleID = resBaru.ID.Hex() // Simpan ID Artikel Baru

		// Insert Artikel Draft
		_, err = repo.Create(ctx, articleDraft)
		require.NoError(t, err)
	})

	t.Run("2. GetAll - Filter Published & Sorting Descending Waktu", func(t *testing.T) {
		// Ambil HANYA yang published
		articles, err := repo.GetAll(ctx, true)

		require.NoError(t, err)
		// Membuktikan filter bekerja: Dari 3 data, hanya 2 yang statusnya Published
		assert.Len(t, articles, 2)

		// Membuktikan Sorting bekerja: Artikel Baru harus muncul duluan (Index 0)
		assert.Equal(t, targetSlug, articles[0].Slug, "Artikel yang lebih baru harus berada di urutan pertama")
		assert.Equal(t, "artikel-kemarin", articles[1].Slug)
	})

	t.Run("3. GetAll - Tanpa Filter (Termasuk Draft)", func(t *testing.T) {
		articles, err := repo.GetAll(ctx, false)

		require.NoError(t, err)
		// Membuktikan tanpa filter mengembalikan semua data
		assert.Len(t, articles, 3)
	})

	t.Run("4. GetBySlug - Sukses Pencarian SEO Friendly", func(t *testing.T) {
		// Kita cari menggunakan targetSlug ("5-tips-koding-golang")
		article, err := repo.GetBySlug(ctx, targetSlug)

		require.NoError(t, err)
		assert.Equal(t, "5 Tips Koding Golang", article.Title)
		// Memastikan data tidak tertukar dengan artikel lain
		assert.Equal(t, true, article.IsPublished)
	})

	t.Run("5. GetBySlug - Gagal Slug Tidak Ada", func(t *testing.T) {
		_, err := repo.GetBySlug(ctx, "slug-hantu")

		assert.Error(t, err)
		assert.Equal(t, mongo.ErrNoDocuments, err)
	})

	t.Run("6. Update - Sukses Mengubah Konten Artikel", func(t *testing.T) {
		updatePayload := domain.Article{
			Title:       "5 Tips Golang (Updated 2026)",
			Slug:        targetSlug, // Anggap slug tetap
			Content:     "Isi sudah diperbarui",
			IsPublished: true,
		}

		updatedArt, err := repo.Update(ctx, createdArticleID, updatePayload)

		require.NoError(t, err)
		// Membuktikan `options.After` bekerja mengembalikan versi final
		assert.Equal(t, "5 Tips Golang (Updated 2026)", updatedArt.Title)
		assert.Equal(t, "Isi sudah diperbarui", updatedArt.Content)
	})

	t.Run("7. Delete - Sukses Menghapus Data", func(t *testing.T) {
		err := repo.Delete(ctx, createdArticleID)
		require.NoError(t, err)

		// Coba cari slug-nya lagi, harusnya sudah hilang
		_, err = repo.GetBySlug(ctx, targetSlug)
		assert.Error(t, err)
		assert.Equal(t, mongo.ErrNoDocuments, err)
	})
}
