package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ConnectMongoDB membuat koneksi ke MongoDB Atlas
func ConnectMongoDB(uri string) *mongo.Client {
	// Connect ke MongoDB (v2 tidak perlu ctx di parameter)
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("❌ Gagal connect ke MongoDB: %v", err)
	}

	// Test koneksi dengan Ping (butuh ctx)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("❌ MongoDB tidak merespon: %v", err)
	}

	log.Println("✅ Berhasil terhubung ke MongoDB!")
	return client
}
