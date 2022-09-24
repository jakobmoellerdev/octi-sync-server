package v1

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

//go:generate oapi-codegen --config REST/oapi-codegen.yaml REST/openapi.yaml
type API struct {
	service.Accounts
	service.Devices
	service.Modules
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
			config.Services.Devices,
			config.Services.Modules,
		},
	}

	auth := api.Group("/auth")
	auth.POST("/register", wrapper.Register)
	auth.POST("/share", wrapper.Share, basicAuthWithShare)

	module := api.Group("/module", basicAuthWithShare)
	module.GET("/:name", wrapper.GetModule)
	module.POST("/:name", wrapper.CreateModule)

	api.GET("/health", wrapper.IsHealthy)
	api.GET("/ready", wrapper.IsReady)
}
