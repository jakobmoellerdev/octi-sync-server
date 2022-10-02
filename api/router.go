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
	DefaultTimeoutSeconds      = 20
	DefaultMaxRequestBodySize  = "64KB"
)

// New generates the api used in the HTTP Server.
func New(ctx context.Context, config *config.Config) http.Handler {
	router := echo.New()

	router.Pre(middleware.RemoveTrailingSlash())

	// CORS Configuration based on config
	router.Use(
		middleware.CORSWithConfig(
			middleware.CORSConfig{
				AllowOrigins: config.Server.CORS.AllowOrigins,
				AllowHeaders: config.Server.CORS.AllowHeaders,
			},
		),
	)

	// XFF Handling for Reverse Proxy Support
	router.IPExtractor = echo.ExtractIPFromXFFHeader()

	// Rate Limiting Support based on simple In-Memory Solution
	router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(RateLimitRequestsPerSecond)))

	// Request ID Tracking for traceability
	router.Use(middleware.RequestID())
	router.Use(
		RequestContextTimeout(config.Server.Timeout.Request),
		MapRequestTimeoutToResponseCode(http.StatusServiceUnavailable),
	)

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

	if config.Server.MaxRequestBodySize == "" {
		config.Server.MaxRequestBodySize = DefaultMaxRequestBodySize
	}
	// Body Size Limitation to avoid Request DOS
	router.Use(middleware.BodyLimit(config.Server.MaxRequestBodySize))

	// Inject V1
	v1.New(ctx, router, config)

	return router
}

func RequestContextTimeout(timeout time.Duration) echo.MiddlewareFunc {
	if timeout == 0 {
		timeout = DefaultTimeoutSeconds * time.Second
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			timeoutCtx, cancel := context.WithTimeout(c.Request().Context(), timeout)

			c.SetRequest(c.Request().WithContext(timeoutCtx))

			defer cancel()

			return next(c)
		}
	}
}

func MapRequestTimeoutToResponseCode(targetCode int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			doneCh := make(chan error)

			run := func(ctx echo.Context) {
				doneCh <- next(ctx)
			}

			go run(ctx)

			select { // wait for task to finish or context to timeout/cancelled
			case err := <-doneCh:
				if err != nil {
					return err
				}

				return nil
			case <-ctx.Request().Context().Done():
				if ctx.Request().Context().Err() == context.DeadlineExceeded {
					return echo.NewHTTPError(targetCode).SetInternal(ctx.Request().Context().Err())
				}

				return ctx.Request().Context().Err() //nolint:wrapcheck
			}
		}
	}
}
