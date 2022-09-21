package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
)

func (api *API) Share(ctx echo.Context, _ ShareParams) error {
	share, err := api.Accounts.Share(ctx.Request().Context(), ctx.Get(auth.UserKey).(string))

	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, &ShareResponse{
		ShareCode: &share,
	})
}
