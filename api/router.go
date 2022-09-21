package api

import (
	"context"
	"net/http"
	"time"

	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/logging"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	RateLimitRequestsPerSecond = 20
	DefaultTimeoutSeconds      = 30
)

// New generates the api used in the HTTP Server.
func New(ctx context.Context, config *config.Config) http.Handler {
	router := echo.New()

	// CORS Configuration based on config
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.Server.CORS.AllowOrigins,
		AllowHeaders: config.Server.CORS.AllowHeaders,
	}))

	// XFF Handling for Reverse Proxy Support
	router.IPExtractor = echo.ExtractIPFromXFFHeader()

	// Rate Limiting Support based on simple In-Memory Solution
	router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(RateLimitRequestsPerSecond)))

	// Request ID Tracking for traceability
	router.Use(middleware.RequestID())

	// Timeout Configuration for timing out requests to 3rd party services
	timeoutConfig := middleware.DefaultTimeoutConfig
	// Default in Middleware is Zero, set to sane default
	timeoutConfig.Timeout = DefaultTimeoutSeconds * time.Second
	router.Use(middleware.TimeoutWithConfig(timeoutConfig))

	// Compression Handlers
	router.Use(
		middleware.Gzip(),
		middleware.Decompress(),
	)

	// Global Middleware for Error Recovery and Request Logging
	router.Use(
		middleware.Recover(),
		logging.RequestLogging(config.Logger),
	)

	// Body Size Limitation to avoid Request DOS
	router.Use(middleware.BodyLimit(config.Server.MaxRequestBodySize))

	// Inject V1
	v1.New(ctx, router, config)

	return router
}
