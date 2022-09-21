package v1

import (
	"fmt"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
	"github.com/labstack/echo/v4"
)

func (api *API) Share(ctx echo.Context, _ ShareParams) error {
	share, err := api.Accounts.Share(ctx.Request().Context(), ctx.Get(auth.UserKey).(string))
	if err != nil {
		return fmt.Errorf("error while attempting to share an account: %w", err)
	}

	if err := ctx.JSON(http.StatusOK, &ShareResponse{
		ShareCode: &share,
	}); err != nil {
		return fmt.Errorf("could not write share response: %w", err)
	}

	return nil
}
