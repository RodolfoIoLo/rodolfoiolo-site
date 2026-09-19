package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/handler"
	appmiddleware "github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/middleware"
)

type Dependencies struct {
	Database handler.Pinger
	Logger   *slog.Logger
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
	e.Use(appmiddleware.AccessLog(dependencies.Logger))

	health := handler.NewHealth(dependencies.Database)
	e.GET("/healthz", health.Liveness)
	e.GET("/readyz", health.Readiness)
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusNotFound) })
	return e
}
