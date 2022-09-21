package v1

import (
	"fmt"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
)

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

	if err := ctx.JSON(aggregation.Health.ToHTTPStatusCode(), &HealthAggregation{
		Components: &components,
		Health:     HealthResult(aggregation.Health),
	}); err != nil {
		return fmt.Errorf("could not write readiness aggregation to response: %w", err)
	}

	return nil
}
