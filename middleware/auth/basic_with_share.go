package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

// UserKey is the cookie name for user credential in basic auth.
const UserKey = "user"

// DeviceID is the cookie name for user credential in basic auth.
const DeviceID = "device"

// DeviceIDHeader holds Device Authentication.
const DeviceIDHeader = "X-Device-ID"

// ShareQueryParamName holds Device Share Codes.
const ShareQueryParamName = "share"

// BasicAuthWithShare returns a Basic HTTP Authorization Handler. It takes as argument a map[string]string where
// the key is the username and the value is the password.
func BasicAuthWithShare(accounts service.Accounts, devices service.Devices) echo.MiddlewareFunc {
	return middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper: middleware.DefaultSkipper,
		Validator: func(username, password string, context echo.Context) (bool, error) {
			ctx := context.Request().Context()
			// Search user in the slice of allowed credentials
			user, err := accounts.Find(ctx, username)
			if err != nil {
				return false, echo.ErrUnauthorized
			}

			if subtle.ConstantTimeCompare([]byte(user.HashedPass()),
				[]byte(fmt.Sprintf("%x", sha256.Sum256([]byte(password))))) != 1 {
				return false, echo.ErrForbidden
			}
			// The user credentials was found, set user's id to key DeviceID in this context,
			// the user's id can be read later using
			// context.MustGet(auth.DeviceID).
			context.Set(UserKey, user)

			deviceID := context.Request().Header.Get(DeviceIDHeader)
			if deviceID == "" {
				return false, echo.NewHTTPError(http.StatusBadRequest,
					"this endpoint has to be called with the "+DeviceIDHeader+" Header!")
			}

			device, err := devices.FindByDeviceID(ctx, user, deviceID)
			if err != nil {
				// Device not found for Account
				share := context.QueryParam(ShareQueryParamName)
				if share, err = url.QueryUnescape(share); err != nil {
					return false, err
				}

				if share == "" {
					// No share Code supplied
					return false, echo.ErrForbidden
				}

				if shared, err := accounts.IsShared(ctx, user.Username(), share); !shared || err != nil {
					// Share code provided but not valid
					return shared, echo.ErrForbidden
				}

				device, err = devices.Register(ctx, user, deviceID)

				if err != nil {
					// Device Registraiton error
					return false, err
				}
			}

			// The account credentials was found, set account's id to key DeviceID in this context,
			// the account's id can be read later using
			// context.MustGet(auth.DeviceID).
			context.Set(DeviceID, device)

			return true, nil
		},
		Realm: "",
	})
}
