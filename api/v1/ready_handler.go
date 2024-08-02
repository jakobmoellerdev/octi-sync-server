package v1

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/jakobmoellerdev/octi-sync-server/api/v1/REST"
	"github.com/jakobmoellerdev/octi-sync-server/service"
)

func (api *API) IsReady(ctx echo.Context) error {
	aggregation := service.HealthAggregator([]service.HealthCheck{
		api.Accounts.HealthCheck(),
		api.Devices.HealthCheck(),
		api.Modules.HealthCheck(),
	}).Check(ctx.Request().Context())

	components := make([]REST.HealthAggregationComponent, len(aggregation.Components))
	for i, component := range aggregation.Components {
		components[i] = REST.HealthAggregationComponent{Health: REST.HealthResult(component.Health), Name: component.Name}
	}

	if err := ctx.JSON(aggregation.Health.ToHTTPStatusCode(), &REST.HealthAggregation{
		Components: &components,
		Health:     REST.HealthResult(aggregation.Health),
	}); err != nil {
		return fmt.Errorf("could not write readiness aggregation to response: %w", err)
	}

	return nil
}
