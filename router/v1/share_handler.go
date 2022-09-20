package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

type (
	ShareHandler struct {
		Accounts service.Accounts
	}

	ShareResponse struct {
		ShareCode string `json:"shareCode" yaml:"shareCode"`
	}
)

func (h *ShareHandler) Share(context echo.Context) error {
	share, err := h.Accounts.Share(context.Request().Context(), context.Get(auth.UserKey).(string))

	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, &ShareResponse{
		ShareCode: share,
	})
}
