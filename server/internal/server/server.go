package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/generated/oapi"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/handler"
	appmiddleware "github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/middleware"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
)

type Dependencies struct {
	Database                             *pgxpool.Pool
	Logger                               *slog.Logger
	API                                  oapi.StrictServerInterface
	AccessVerifier                       appmiddleware.AccessVerifier
	VisitorSigner                        security.VisitorSigner
	RefreshCookieName, VisitorCookieName string
	CookieSecure                         bool
	VisitorSecret                        []byte
	AllowedOrigins                       []string
	UploadRoot                           string
}

func New(dependencies Dependencies) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Server.ReadHeaderTimeout = 5 * time.Second
	e.Server.ReadTimeout = 15 * time.Second
	e.Server.WriteTimeout = 30 * time.Second
	e.Server.IdleTimeout = 60 * time.Second
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.BodyLimit("7M"))
	e.Use(appmiddleware.SecurityHeaders)
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{AllowOrigins: dependencies.AllowedOrigins, AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}, AllowHeaders: []string{echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderContentType, echo.HeaderOrigin, "X-Requested-With"}, AllowCredentials: true, MaxAge: 600}))
	e.Use(appmiddleware.AuthContext(dependencies.AccessVerifier, dependencies.RefreshCookieName, dependencies.VisitorSecret))
	e.Use(appmiddleware.VisitorIdentity(dependencies.VisitorCookieName, dependencies.CookieSecure, dependencies.VisitorSigner))
	e.Use(appmiddleware.CSRF(dependencies.AllowedOrigins))
	e.Use(appmiddleware.NewLoginLimiter().Middleware)
	e.Use(appmiddleware.AccessLog(dependencies.Logger))

	strict := oapi.NewStrictHandler(dependencies.API, nil)
	oapi.RegisterHandlersWithBaseURL(e, strict, "/api/v1")
	health := handler.NewHealth(dependencies.Database)
	e.GET("/healthz", health.Liveness)
	e.GET("/readyz", health.Readiness)
	if dependencies.UploadRoot != "" {
		e.Static("/uploads", dependencies.UploadRoot)
	}
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusNotFound) })
	return e
}
