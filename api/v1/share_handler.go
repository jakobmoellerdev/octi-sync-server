package v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jakobmoellerdev/octi-sync-server/api/v1/REST"
	"github.com/jakobmoellerdev/octi-sync-server/middleware/basic"
	"github.com/jakobmoellerdev/octi-sync-server/service"
)

func (api *API) Share(ctx echo.Context, _ REST.ShareParams) error {
	account, found := ctx.Get(basic.AccountKey).(service.Account)
	if !found {
		return echo.ErrForbidden
	}

	share, err := api.Sharing.Share(ctx.Request().Context(), account)
	if err != nil {
		return fmt.Errorf("error while attempting to share an account: %w", err)
	}

	if err := ctx.JSON(http.StatusOK, &REST.ShareResponse{
		ShareCode: (*string)(&share),
	}); err != nil {
		return fmt.Errorf("could not write share response: %w", err)
	}

	return nil
}
