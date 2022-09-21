package v1

import (
	"context"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type API struct {
	service.Accounts
	service.Devices
	service.Modules
}

const Prefix = "/v1"

func New(_ context.Context, engine *echo.Echo, config *config.Config) {
	v1 := engine.Group(Prefix) //nolint:varnamelen

	swagger, err := GetSwagger()
	if err != nil {
		config.Logger.Fatal().Err(err)
	}
	serversToLog := zerolog.Arr()
	for _, server := range swagger.Servers {
		serversToLog = serversToLog.Str(server.URL)
	}
	config.Logger.Info().
		Str("api", swagger.Info.Title).
		Str("version", swagger.Info.Version).
		Array("servers", serversToLog).
		Msg("API Loaded!")
	swaggerJSON, err := swagger.MarshalJSON()
	if err != nil {
		config.Logger.Fatal().Err(err)
	}
	v1.GET("/openapi", func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, swaggerJSON)
	})

	middleware := auth.BasicAuthWithShare(config.Services.Accounts, config.Services.Devices)
	v1.Group("/auth/share").Use(middleware)
	v1.Group("/module").Use(middleware)

	RegisterHandlers(v1, &API{
		config.Services.Accounts,
		config.Services.Devices,
		config.Services.Modules,
	})
}

func (api *API) IsHealthy(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, &HealthAggregation{Health: Up})
}

func (api *API) IsReady(ctx echo.Context) error {
	aggregation := service.HealthAggregator([]service.HealthCheck{
		api.Accounts.HealthCheck(),
		api.Devices.HealthCheck(),
		api.Modules.HealthCheck(),
	}).Check(ctx.Request().Context())

	components := make([]HealthAggregationComponent, len(aggregation.Components))
	for i, component := range aggregation.Components {
		components[i] = HealthAggregationComponent{HealthResult(component.Health), component.Name}
	}
	return ctx.JSON(aggregation.Health.ToHTTPStatusCode(), &HealthAggregation{
		Components: &components,
		Health:     HealthResult(aggregation.Health),
	})
}
