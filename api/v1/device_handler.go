package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
)

var ErrNoDeviceAccessWithoutAccount = echo.NewHTTPError(http.StatusForbidden,
	errors.New("devices cannot be accessed without an account"))

func (api *API) GetDevices(ctx echo.Context, _ REST.GetDevicesParams) error {
	account, found := ctx.Get(basic.AccountKey).(service.Account)
	if !found {
		return ErrNoDeviceAccessWithoutAccount
	}

	devicesFromAccount, err := api.Devices.FindByAccount(
		ctx.Request().Context(),
		account,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("could not fetch devices from account: %w", err))
	}

	devices := make([]REST.Device, len(devicesFromAccount))
	for i, device := range devicesFromAccount {
		devices[i] = REST.Device{Id: REST.DeviceID(device.ID())}
	}

	if err := ctx.JSON(http.StatusOK, &REST.DeviceListResponse{
		Count: len(devicesFromAccount),
		Items: devices,
	}); err != nil {
		return fmt.Errorf("could not write device list response: %w", err)
	}

	return nil
}
