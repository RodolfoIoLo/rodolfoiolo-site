package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/config"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/database"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/handler"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	appserver "github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/server"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/service"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	cfg, err := config.Load()
	if err != nil {
		fatal(logger, "load_configuration", err)
	}
	rootContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := database.Open(rootContext, database.Options{URL: cfg.DatabaseURL, MaxConns: cfg.DatabaseMaxConns, MinConns: cfg.DatabaseMinConns, Timeout: cfg.DatabaseTimeout})
	if err != nil {
		fatal(logger, "connect_database", err)
	}
	defer pool.Close()
	jwtManager, err := security.LoadJWTManager(cfg.JWTPrivateKeyFile, cfg.JWTPublicKeyFile, cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTTL)
	if err != nil {
		fatal(logger, "load_jwt", err)
	}
	secretEncoded, err := os.ReadFile(cfg.VisitorSecretFile)
	if err != nil {
		fatal(logger, "read_visitor_secret", err)
	}
	visitorSecret, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(string(secretEncoded)))
	if err != nil {
		fatal(logger, "decode_visitor_secret", err)
	}
	visitorSigner, err := security.NewVisitorSigner(visitorSecret)
	if err != nil {
		fatal(logger, "create_visitor_signer", err)
	}
	objectStore, err := storage.NewLocal(cfg.UploadRoot, cfg.MediaPublicURL)
	if err != nil {
		fatal(logger, "create_object_store", err)
	}
	authRepository := repository.NewAuthRepository(pool)
	categoryRepository := repository.NewCategoryRepository(pool)
	tagRepository := repository.NewTagRepository(pool)
	articleRepository := repository.NewArticleRepository(pool)
	projectRepository := repository.NewProjectRepository(pool)
	siteRepository := repository.NewSiteRepository(pool)
	interactionRepository := repository.NewInteractionRepository(pool)
	mediaRepository := repository.NewMediaRepository(pool)
	authService := service.NewAuthService(authRepository, jwtManager, cfg.RefreshTTL)
	contentService := service.NewContentService(categoryRepository, tagRepository, articleRepository, projectRepository, siteRepository)
	interactionService := service.NewInteractionService(interactionRepository, security.NewCommentSanitizer(), visitorSecret)
	mediaService := service.NewMediaService(mediaRepository, objectStore)
	api := handler.New(handler.Dependencies{Auth: authService, Cookies: security.CookieManager{Name: cfg.RefreshCookieName, Secure: cfg.CookieSecure, TTL: cfg.RefreshTTL}, Content: contentService, Interactions: interactionService, Media: mediaService, Database: pool, PublicURL: cfg.PublicSiteURL})
	application := appserver.New(appserver.Dependencies{Database: pool, Logger: logger, API: api, AccessVerifier: jwtManager, VisitorSigner: visitorSigner, RefreshCookieName: cfg.RefreshCookieName, VisitorCookieName: cfg.VisitorCookieName, CookieSecure: cfg.CookieSecure, VisitorSecret: visitorSecret, AllowedOrigins: cfg.AllowedOrigins, UploadRoot: cfg.UploadRoot})
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
		fatal(logger, "graceful_shutdown_failed", err)
	}
	logger.Info("api_stopped")
}
func fatal(logger *slog.Logger, message string, err error) {
	logger.Error(message, "error", err)
	os.Exit(1)
}
