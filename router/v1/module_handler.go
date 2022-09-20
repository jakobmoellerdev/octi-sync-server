package v1

import (
	"fmt"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/redis"
	"github.com/labstack/echo/v4"
)

type ModuleHandler struct {
	modules service.Modules
}

func (h *ModuleHandler) CreateModule() echo.HandlerFunc {
	return func(context echo.Context) error {
		moduleName := context.Param("name")
		if moduleName == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "module name is required")
		}

		deviceID, ok := context.Get(auth.DeviceID).(string)
		if !ok || deviceID == "" {
			return ErrDeviceIDNotPropagated
		}

		err := h.modules.Set(
			context.Request().Context(),
			fmt.Sprintf("%s-%s", deviceID, moduleName),
			redis.ModuleFromReader(context.Request().Body, int(context.Request().ContentLength)),
		)
		if err != nil {
			return err
		}

		return context.JSON(http.StatusOK, nil)
	}
}

func (h *ModuleHandler) GetModule() echo.HandlerFunc {
	return func(context echo.Context) error {
		moduleName := context.Param("name")
		if moduleName == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "module name is required")
		}

		deviceID, ok := context.Get(auth.DeviceID).(string)
		if !ok || deviceID == "" {
			return ErrDeviceIDNotPropagated
		}

		module, err := h.modules.Get(context.Request().Context(),
			fmt.Sprintf("%s-%s", deviceID, moduleName))
		if err != nil {
			return err
		}

		return context.Stream(http.StatusOK, echo.MIMEOctetStream, module.Raw())
	}
}
