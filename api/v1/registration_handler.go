package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
)

func (api *API) Register(ctx echo.Context, params RegisterParams) error {
	accountID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	deviceID := params.XDeviceID
	if deviceID == "" {
		deviceUUID, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		deviceID = deviceUUID.String()
		ctx.Response().Header().Set(auth.DeviceIDHeader, deviceID)
	}

	acc, password, err := api.Accounts.Register(ctx.Request().Context(), accountID.String())
	if err != nil {
		return err
	}

	if _, err := api.Devices.Register(ctx.Request().Context(), acc, deviceID); err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, &RegistrationResponse{
		Username: acc.Username(),
		DeviceID: deviceID,
		Password: password,
	})
}
