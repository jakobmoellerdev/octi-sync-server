package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/redis"
	"github.com/labstack/echo/v4"
)

var ErrAccountForVerifyingDeviceNotPresent = errors.New("account for verifying device id in params is not present")

func (api *API) CreateModule(ctx echo.Context, name REST.ModuleName, params REST.CreateModuleParams) error {
	err := api.Modules.Set(
		ctx.Request().Context(),
		fmt.Sprintf("%s-%s", params.XDeviceID, name),
		redis.ModuleFromReader(ctx.Request().Body, int(ctx.Request().ContentLength)),
	)
	if err != nil {
		return fmt.Errorf("could not create/update module: %w", err)
	}

	if err := ctx.JSON(http.StatusAccepted, nil); err != nil {
		return fmt.Errorf("could not acknowledge module creation: %w", err)
	}

	return nil
}

func (api *API) GetModule(ctx echo.Context, name REST.ModuleName, params REST.GetModuleParams) error {
	deviceID := params.XDeviceID

	if params.DeviceId != nil {
		deviceID = *params.DeviceId
		acc, accountPresent := ctx.Get(basic.AccountKey).(service.Account)

		if !accountPresent {
			return echo.NewHTTPError(
				http.StatusForbidden, ErrAccountForVerifyingDeviceNotPresent,
			)
		}

		if _, err := api.GetDevice(ctx.Request().Context(), acc, service.DeviceID(deviceID)); err != nil {
			return echo.NewHTTPError(
				http.StatusForbidden, fmt.Errorf(
					"device from params could not be verified against account: %w", err,
				),
			)
		}
	}

	module, err := api.Modules.Get(
		ctx.Request().Context(),
		fmt.Sprintf("%s-%s", deviceID, name),
	)
	if err != nil {
		return fmt.Errorf("error while fetching module: %w", err)
	}

	if err := ctx.Stream(http.StatusOK, echo.MIMEOctetStream, module.Raw()); err != nil {
		return fmt.Errorf("error while writing module data to response while fetching module: %w", err)
	}

	return nil
}
