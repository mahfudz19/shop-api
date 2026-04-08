package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB(uri string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Gagal menyambung ke MongoDB: %v", err)
	}

	// Tes Ping koneksi
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB tidak merespon: %v", err)
	}

	log.Println("Berhasil terhubung ke MongoDB!")
	return client
}