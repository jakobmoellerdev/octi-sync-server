package v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jakob-moeller-cloud/octi-sync-server/service/redis"
)

func (api *API) CreateModule(ctx echo.Context, name ModuleName, params CreateModuleParams) error {
	err := api.Modules.Set(
		ctx.Request().Context(),
		fmt.Sprintf("%s-%s", params.XDeviceID, name),
		redis.ModuleFromReader(ctx.Request().Body, int(ctx.Request().ContentLength)),
	)

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, nil)
}

func (api *API) GetModule(ctx echo.Context, name ModuleName, params GetModuleParams) error {
	module, err := api.Modules.Get(ctx.Request().Context(),
		fmt.Sprintf("%s-%s", params.XDeviceID, name))
	if err != nil {
		return err
	}

	return ctx.Stream(http.StatusOK, echo.MIMEOctetStream, module.Raw())
}
