package config

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis menginisialisasi koneksi ke server Redis
func ConnectRedis() *redis.Client {
	// 1. Ambil nilai konfigurasi dari Environment Variables
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		// Nilai fallback (cadangan) jika file .env kosong / tidak terbaca
		redisHost = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	// 2. Ambil nilai DB (opsional), jadikan integer. Default adalah 0.
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if dbInt, err := strconv.Atoi(dbStr); err == nil {
			redisDB = dbInt
		}
	}

	// 3. Masukkan variabel tersebut ke konfigurasi Redis
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Uji koneksi (Ping)
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		// Menambahkan informasi 'redisHost' di pesan error untuk mempermudah debugging
		log.Fatalf("❌ Gagal terhubung ke Redis di [%s]: %v", redisHost, err)
	}

	log.Println("✅ Berhasil terhubung ke Redis Queue")
	return client
}
