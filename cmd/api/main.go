package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/assidik12/catalyst/cmd/injector"
	"github.com/assidik12/catalyst/config"
	"github.com/assidik12/catalyst/internal/infrastructure"
	"github.com/assidik12/catalyst/internal/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 1. Load Config
	cfg := config.GetConfig()

	// 2. Initialize Logger
	l := logger.New(cfg.AppEnv)
	slog.SetDefault(l)

	// Initialize OpenTelemetry Tracer
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:14268/api/traces"
	}
	tp, err := infrastructure.InitTracer("catalyst-backend", jaegerEndpoint)
	if err != nil {
		l.Error("Failed to initialize tracer", "error", err)
	} else {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				l.Error("Error shutting down tracer provider", "error", err)
			}
		}()
	}

	// 3. Initialize Server via Wire
	app, cleanup, err := injector.InitializedServer(*cfg)
	if err != nil {
		l.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}

	// 4. Cleanup resources (Close DB/Redis/Kafka connections) when app exits
	if cleanup != nil {
		defer cleanup()
	}

	app.Server.Addr = fmt.Sprintf(":%s", cfg.AppPort)

	relayCtx, cancelRelay := context.WithCancel(context.Background())
	defer cancelRelay()

	// Start Relay Worker
	go app.Relay.Start(relayCtx)

	// 5. Start server in a goroutine
	go func() {
		l.Info("Server starting", "port", cfg.AppPort)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// 6. Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down server gracefully...")

	// 7. Context for shutdown with 30s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		l.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	l.Info("Server exited cleanly")
}
