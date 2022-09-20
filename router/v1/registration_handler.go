package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
)

type (
	RegistrationHandler struct {
		service.Accounts
		service.Devices
	}

	RegistrationResponse struct {
		DeviceID string `json:"deviceId" yaml:"deviceId"`
		Username string `json:"username" yaml:"username"`
		Password string `json:"password" yaml:"password"`
	}
)

func (h *RegistrationHandler) Register(context echo.Context) error {
	accountID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	deviceID := context.Request().Header.Get(auth.DeviceIDHeader)
	if deviceID == "" {
		deviceUUID, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		deviceID = deviceUUID.String()
		context.Response().Header().Set(auth.DeviceIDHeader, deviceID)
	}

	acc, password, err := h.Accounts.Register(context.Request().Context(), accountID.String())
	if err != nil {
		return err
	}

	if _, err := h.Devices.Register(context.Request().Context(), acc, deviceID); err != nil {
		return err
	}

	return context.JSON(http.StatusOK, &RegistrationResponse{
		Username: acc.Username(),
		DeviceID: deviceID,
		Password: password,
	})
}
