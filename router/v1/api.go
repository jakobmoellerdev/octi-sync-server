package v1

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	authmiddleware "github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
)

var ErrDeviceIDNotPropagated = echo.NewHTTPError(http.StatusInternalServerError,
	"device id was not propagated from middleware")

func New(_ context.Context, engine *echo.Echo, config *config.Config) {
	v1 := engine.Group("/v1") //nolint:varnamelen

	{
		middleware := authmiddleware.BasicAuthWithShare(config.Services.Accounts, config.Services.Devices)
		auth := v1.Group("/auth")
		registrationHandler := &RegistrationHandler{
			config.Services.Accounts,
			config.Services.Devices,
		}
		shareHandler := &ShareHandler{
			config.Services.Accounts,
		}
		{
			auth.POST("/register", registrationHandler.Register)
			auth.GET("/share", shareHandler.Share, middleware)
		}

		module := v1.Group("/module", middleware)
		moduleHandler := &ModuleHandler{
			config.Services.Modules,
		}
		{
			module.GET("/:name", moduleHandler.GetModule())
			module.POST("/:name", moduleHandler.CreateModule())
		}
	}
}
