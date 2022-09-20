package router

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/logging"
	v1 "github.com/jakob-moeller-cloud/octi-sync-server/router/v1"
)

const (
	RateLimitRequestsPerSecond = 20
	DefaultTimeoutSeconds      = 30
)

// New generates the router used in the HTTP Server.
func New(ctx context.Context, config *config.Config) http.Handler {
	router := echo.New()

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

	router.GET("/ready", ReadyCheck(config))
	router.GET("/health", HealthCheck)

	return router
}

func ReadyCheck(cfg *config.Config) echo.HandlerFunc {
	return func(context echo.Context) error {
		aggregation := service.HealthAggregator([]service.HealthCheck{
			cfg.Services.Accounts.HealthCheck(),
			cfg.Services.Devices.HealthCheck(),
			cfg.Services.Modules.HealthCheck(),
		}).Check(context.Request().Context())

		if aggregation.Health == service.HealthUp {
			return context.JSON(http.StatusOK, aggregation)
		}
		return context.JSON(http.StatusServiceUnavailable, aggregation)
	}
}

func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, struct{}{})
}
