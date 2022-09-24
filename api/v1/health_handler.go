package v1

import (
	"fmt"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/labstack/echo/v4"
)

func (api *API) IsHealthy(ctx echo.Context) error {
	if err := ctx.JSON(http.StatusOK, &REST.HealthAggregation{Health: REST.Up}); err != nil {
		return fmt.Errorf("could not write healthiness aggregation to response: %w", err)
	}

	return nil
}
