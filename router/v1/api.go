package v1

import (
	"context"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	authmiddleware "github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
	"github.com/labstack/echo/v4"
)

var ErrDeviceIDNotPropagated = echo.NewHTTPError(http.StatusInternalServerError,
	"device id was not propagated from middleware")

func New(_ context.Context, engine *echo.Echo, config *config.Config) {
	v1 := engine.Group("/v1") //nolint:varnamelen

	{
		auth := v1.Group("/auth")
		registrationHandler := &RegistrationHandler{
			config.Services.Accounts,
			config.Services.Devices,
		}
		{
			auth.POST("/Register", registrationHandler.Register)
		}

		module := v1.Group("/module")
		module.Use(
			authmiddleware.BasicAuthWithShare(config.Services.Accounts, config.Services.Devices),
		)
		moduleHandler := &ModuleHandler{
			config.Services.Modules,
		}
		{
			module.GET("/:name", moduleHandler.GetModule())
			module.POST("/:name", moduleHandler.CreateModule())
		}
	}
}
