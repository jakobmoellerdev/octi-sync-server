package router

import (
	"context"
	"fmt"
	"go.jakob-moeller.cloud/octi-sync-server/service"
	"net/http"

	"go.jakob-moeller.cloud/octi-sync-server/config"
	"go.jakob-moeller.cloud/octi-sync-server/middleware/logging"
	requestmiddleware "go.jakob-moeller.cloud/octi-sync-server/middleware/request"
	v1 "go.jakob-moeller.cloud/octi-sync-server/router/v1"

	"github.com/gin-contrib/gzip"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
)

// New generates the router used in the HTTP Server.
func New(ctx context.Context, config *config.Config) http.Handler {
	router := gin.New()

	router.Use(requestmiddleware.LimitHandler(requestmiddleware.DefaultLimit()))
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Global Middleware
	router.Use(
		ginzap.RecoveryWithZap(config.Logger, true),
		logging.RequestLogging(config.Logger),
	)

	router.Use(requestmiddleware.BodySizeLimiter(config.Server.MaxRequestBodySize, gin.H{
		"msg": fmt.Sprintf("request too large, maximum allowed is %v bytes", config.Server.MaxRequestBodySize),
	}))

	v1.New(ctx, router, config)

	router.GET("/ready", healthCheck(config))
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return router
}

func healthCheck(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		aggregation := service.HealthAggregator([]service.HealthCheck{
			cfg.Services.Accounts.HealthCheck(),
			cfg.Services.Devices.HealthCheck(),
			cfg.Services.Modules.HealthCheck(),
		}).Check(c.Request.Context())

		if aggregation.Health == service.HealthUp {
			c.JSON(http.StatusOK, aggregation)
		} else {
			c.JSON(http.StatusServiceUnavailable, aggregation)
		}
	}
}
