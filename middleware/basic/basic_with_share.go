package basic

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/jakobmoellerdev/octi-sync-server/service"
)

// AccountKey is the cookie name for user credential in basic auth.
const AccountKey = "user"

// Device is the cookie name for user credential in basic auth.
const Device = "device"

// DeviceIDHeader holds Device Authentication.
const DeviceIDHeader = "X-Device-ID"

var ErrDevicePassVerificationFailed = errors.New("device pass verification failed")

// AuthWithShare returns a Basic HTTP Authorization Handler. It takes as argument a map[string]string where
// the key is the username and the value is the password.
func AuthWithShare(accounts service.Accounts, devices service.Devices) echo.MiddlewareFunc {
	return middleware.BasicAuthWithConfig(
		middleware.BasicAuthConfig{
			Skipper: middleware.DefaultSkipper,
			Validator: func(username, password string, context echo.Context) (bool, error) {
				ctx := context.Request().Context()
				// Search account in the slice of allowed credentials
				account, err := accounts.Find(ctx, username)
				if err != nil {
					return false, echo.ErrUnauthorized
				}

				// The account credentials was found, set account's id to key Device in this context,
				// the account's id can be read later using
				// context.MustGet(auth.AccountKey).
				context.Set(AccountKey, account)

				deviceIDFromHeader := context.Request().Header.Get(DeviceIDHeader)
				if deviceIDFromHeader == "" {
					return false, echo.NewHTTPError(
						http.StatusBadRequest,
						"this endpoint has to be called with the "+DeviceIDHeader+" Header!",
					)
				}
				deviceID, err := uuid.Parse(deviceIDFromHeader)
				if err != nil {
					return false, echo.NewHTTPError(
						http.StatusBadRequest,
						DeviceIDHeader+" has to be a valid UUID!",
					)
				}

				device, err := devices.GetDevice(ctx, account, service.DeviceID(deviceID))

				if errors.Is(err, service.ErrDeviceNotFound) {
					return false, echo.NewHTTPError(http.StatusForbidden).SetInternal(err)
				}

				if !device.Verify(password) {
					return false, echo.ErrForbidden.SetInternal(ErrDevicePassVerificationFailed)
				}

				// The account credentials was found, set account's id to key Device in this context,
				// the account's id can be read later using
				// context.MustGet(auth.Device).
				context.Set(Device, device)

				return true, nil
			},
			Realm: "",
		},
	)
}
