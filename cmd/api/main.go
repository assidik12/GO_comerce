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

	cfg := config.GetConfig()

	server := injector.InitializedServer(*cfg)

	server.Addr = fmt.Sprintf(":%s", cfg.AppPort)

	log.Printf("Server is starting on port %s...", cfg.AppPort)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
