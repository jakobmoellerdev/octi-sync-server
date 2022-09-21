package v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (api *API) IsHealthy(ctx echo.Context) error {
	if err := ctx.JSON(http.StatusOK, &HealthAggregation{Health: Up}); err != nil {
		return fmt.Errorf("could not write healthiness aggregation to response: %w", err)
	}

	return nil
}
