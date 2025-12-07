package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/assidik12/go-restfull-api/cmd/injector"
	"github.com/assidik12/go-restfull-api/config"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 1. Load Config
	cfg := config.GetConfig()

	// 2. Initialize Server via Wire
	// Perhatikan: Argument kedua hilang. Redis & Kafka otomatis di-inject di dalam.
	// Kita juga menerima fungsi cleanup untuk menutup koneksi database/redis dengan anggun.
	server, cleanup, err := injector.InitializedServer(*cfg)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// 3. Cleanup resources (Close DB/Redis/Kafka connections) saat aplikasi mati
	if cleanup != nil {
		defer cleanup()
	}

	server.Addr = fmt.Sprintf(":%s", cfg.AppPort)

	log.Printf("Server GO_comerce is starting on port %s...", cfg.AppPort)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
