package api

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/logging"
)

const (
	RateLimitRequestsPerSecond = 20
	DefaultTimeoutSeconds      = 30
)

// New generates the api used in the HTTP Server.
func New(ctx context.Context, config *config.Config) http.Handler {
	router := echo.New()

	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.Server.CORS.AllowOrigins,
		AllowHeaders: config.Server.CORS.AllowHeaders,
	}))

	router.IPExtractor = echo.ExtractIPFromXFFHeader()

	router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(RateLimitRequestsPerSecond)))

	router.Use(middleware.RequestID())

	timeoutConfig := middleware.DefaultTimeoutConfig
	// Default in Middleware is Zero, set to sane default
	timeoutConfig.Timeout = DefaultTimeoutSeconds * time.Second
	router.Use(middleware.TimeoutWithConfig(timeoutConfig))

	router.Use(
		middleware.Gzip(),
		middleware.Decompress(),
	)

	// Global Middleware
	router.Use(
		middleware.Recover(),
		logging.RequestLogging(config.Logger),
	)

	router.Use(middleware.BodyLimit(config.Server.MaxRequestBodySize))

	v1.New(ctx, router, config)

	return router
}
