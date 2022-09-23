package v1

import (
	"context"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
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

	middleware := auth.BasicAuthWithShare(config.Services.Accounts, config.Services.Devices)
	api.Group("/auth/share").Use(middleware)
	api.Group("/module").Use(middleware)

	REST.RegisterHandlers(api, &API{
		config.Services.Accounts,
		config.Services.Devices,
		config.Services.Modules,
	})
}
