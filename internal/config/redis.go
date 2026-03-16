package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis menginisialisasi koneksi ke server Redis
func ConnectRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Sesuaikan jika Redis ada di server lain
		Password: "",               // Kosongkan jika tidak ada password
		DB:       0,                // Database default Redis
	})

	// Uji koneksi (Ping)
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Gagal terhubung ke Redis: %v", err)
	}

	log.Println("✅ Berhasil terhubung ke Redis Queue")
	return client
}
