package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/config"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/database"
	appserver "github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load_configuration", "error", err)
		os.Exit(1)
	}

	rootContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Open(rootContext, database.Options{
		URL:      cfg.DatabaseURL,
		MaxConns: cfg.DatabaseMaxConns,
		MinConns: cfg.DatabaseMinConns,
		Timeout:  cfg.DatabaseTimeout,
	})
	if err != nil {
		logger.Error("connect_database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	application := appserver.New(appserver.Dependencies{Database: pool, Logger: logger})
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("api_starting", "address", cfg.HTTPAddress, "environment", cfg.Environment)
		serverErrors <- application.Start(cfg.HTTPAddress)
	}()

	select {
	case <-rootContext.Done():
		logger.Info("shutdown_requested")
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http_server_failed", "error", err)
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := application.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful_shutdown_failed", "error", err)
		os.Exit(1)
	}
	logger.Info("api_stopped")
}
