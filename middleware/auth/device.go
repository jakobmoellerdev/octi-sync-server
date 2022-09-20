package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

// DeviceID is the cookie name for user credential in basic auth.
const DeviceID = "device"

// DeviceIDHeader holds Device Authentication.
const DeviceIDHeader = "X-Device-ID"

// DeviceAuth returns a Basic HTTP Authorization middleware. It takes as arguments a map[string]string where
// the key is the username and the value is the password, as well as the name of the Realm.
// If the realm is empty, "Authorization Required" will be used by default.
// (see http://tools.ietf.org/html/rfc2617#section-1.2)
func DeviceAuth(devices service.Devices, skipper middleware.Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			if skipper(context) {
				return next(context)
			}
			account, ok := context.Get(UserKey).(service.Account)
			if !ok || account == nil {
				// Account not found, we return 401 and abort handlers chain.
				return echo.ErrUnauthorized
			}

			deviceID := context.Request().Header.Get(DeviceIDHeader)
			if deviceID == "" {
				return echo.NewHTTPError(http.StatusBadRequest,
					"this endpoint has to be called with the "+DeviceIDHeader+" Header!")
			}

			device, err := devices.FindByDeviceID(
				context.Request().Context(),
				account, deviceID)
			if err != nil {
				// Device not found for Account, we return 401 and abort handlers chain.
				return echo.ErrUnauthorized
			}

			// The account credentials was found, set account's id to key DeviceID in this context,
			// the account's id can be read later using
			// context.MustGet(auth.DeviceID).
			context.Set(DeviceID, device)

			return next(context)
		}
	}
}
