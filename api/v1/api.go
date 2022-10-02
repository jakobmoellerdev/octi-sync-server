package v1

import (
	"context"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-password/password"
)

//go:generate oapi-codegen --config REST/oapi-codegen.yaml REST/openapi.yaml
type API struct {
	service.Accounts
	service.Sharing
	service.Devices
	service.Modules
	service.MetadataProvider
	password.PasswordGenerator
	service.UsernameGenerator
}

const Prefix = "/v1"

func New(_ context.Context, engine *echo.Echo, config *config.Config) {
	api := engine.Group(Prefix)

	swagger, err := REST.GetSwagger()
	if err != nil {
		config.Logger.Fatal().Err(err).Msg("error while resolving swagger")
	}

	api.GET("/openapi", NewOpenAPIHandler(swagger, config.Logger).ServeOpenAPI)

	basicAuthWithShare := basic.AuthWithShare(config.Services.Accounts, config.Services.Devices)

	wrapper := REST.ServerInterfaceWrapper{
		Handler: &API{
			config.Services.Accounts,
			config.Services.Sharing,
			config.Services.Devices,
			config.Services.Modules,
			config.Services.MetadataProvider,
			config.PasswordGenerator,
			config.UsernameGenerator,
		},
	}

	auth := api.Group("/auth")
	auth.POST("/register", wrapper.Register)
	auth.POST("/share", wrapper.Share, basicAuthWithShare)

	module := api.Group("/module", basicAuthWithShare)
	module.GET("/:name", wrapper.GetModule)
	module.POST("/:name", wrapper.CreateModule)
	module.DELETE("", wrapper.DeleteModules)

	api.GET("/devices", wrapper.GetDevices, basicAuthWithShare)

	api.GET("/health", wrapper.IsHealthy)
	api.GET("/ready", wrapper.IsReady)
}
