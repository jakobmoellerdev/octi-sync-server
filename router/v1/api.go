package v1

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"

	authmiddleware "github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"

	"github.com/google/uuid"
)

var ErrDeviceIDNotPropagated = echo.NewHTTPError(http.StatusInternalServerError,
	"device id was not propagated from middleware")

func New(_ context.Context, engine *echo.Echo, config *config.Config) {
	v1 := engine.Group("/v1") //nolint:varnamelen

	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", register(
				config.Services.Accounts,
				config.Services.Devices,
			))
		}

		module := v1.Group("/module")
		module.Use(
			authmiddleware.BasicAuth(config.Services.Accounts),
			authmiddleware.DeviceAuth(config.Services.Devices, middleware.DefaultSkipper),
		)
		{
			module.GET("/:name", getModule(config.Services.Modules))
			module.POST("/:name", createModule(config.Services.Modules))
		}
	}
}

func createModule(modules service.Modules) echo.HandlerFunc {
	return func(context echo.Context) error {
		moduleName := context.Param("name")
		if moduleName == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "module name is required")
		}

		deviceID, ok := context.Get(authmiddleware.DeviceID).(string)
		if !ok || deviceID == "" {
			return ErrDeviceIDNotPropagated
		}

		err := modules.Set(
			context.Request().Context(),
			fmt.Sprintf("%s-%s", deviceID, moduleName),
			service.RedisModuleFromReader(context.Request().Body, int(context.Request().ContentLength)),
		)
		if err != nil {
			return err
		}

		return context.JSON(http.StatusOK, nil)
	}
}

func getModule(modules service.Modules) echo.HandlerFunc {
	return func(context echo.Context) error {
		moduleName := context.Param("name")
		if moduleName == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "module name is required")
		}

		deviceID, ok := context.Get(authmiddleware.DeviceID).(string)
		if !ok || deviceID == "" {
			return ErrDeviceIDNotPropagated
		}

		module, err := modules.Get(context.Request().Context(),
			fmt.Sprintf("%s-%s", deviceID, moduleName))
		if err != nil {
			return err
		}

		// context.Stream(func(w io.Writer) bool {
		//	if _, err := io.Copy(w, module.Raw()); err != nil {
		//		_ = context.AbortWithError(http.StatusInternalServerError, err)
		//	}
		//	return false
		// })

		return context.Stream(http.StatusOK, "application/octet-stream", module.Raw())
	}
}

type RegistrationResponse struct {
	Username string
	DeviceID string
	Password string
}

func register(accounts service.Accounts, devices service.Devices) echo.HandlerFunc {
	return func(context echo.Context) error {
		accountID, err := uuid.NewRandom()
		if err != nil {
			return err
		}

		deviceID := context.Request().Header.Get(authmiddleware.DeviceIDHeader)
		if deviceID == "" {
			deviceUUID, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			deviceID = deviceUUID.String()
			context.Response().Header().Set(authmiddleware.DeviceIDHeader, deviceID)
		}

		acc, password, err := accounts.Register(context.Request().Context(), accountID.String())
		if err != nil {
			return err
		}

		if err := devices.Register(context.Request().Context(), acc, deviceID); err != nil {
			return err
		}

		return context.JSON(http.StatusOK, &RegistrationResponse{
			Username: acc.Username(),
			DeviceID: deviceID,
			Password: password,
		})
	}
}
