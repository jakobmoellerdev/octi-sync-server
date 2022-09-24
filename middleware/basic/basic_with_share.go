package basic

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// AccountKey is the cookie name for user credential in basic auth.
const AccountKey = "user"

// DeviceID is the cookie name for user credential in basic auth.
const DeviceID = "device"

// DeviceIDHeader holds Device Authentication.
const DeviceIDHeader = "X-Device-ID"

// ShareQueryParamName holds Device Share Codes.
const ShareQueryParamName = "share"

// AuthWithShare returns a Basic HTTP Authorization Handler. It takes as argument a map[string]string where
// the key is the username and the value is the password.
func AuthWithShare(accounts service.Accounts, devices service.Devices) echo.MiddlewareFunc {
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
			context.Set(AccountKey, user)

			deviceIDFromHeader := context.Request().Header.Get(DeviceIDHeader)
			if deviceIDFromHeader == "" {
				return false, echo.NewHTTPError(http.StatusBadRequest,
					"this endpoint has to be called with the "+DeviceIDHeader+" Header!")
			}
			deviceID, err := uuid.Parse(deviceIDFromHeader)
			if err != nil {
				return false, echo.NewHTTPError(http.StatusBadRequest,
					DeviceIDHeader+" has to be a valid UUID!")
			}

			device, err := devices.FindByDeviceID(ctx, user, service.DeviceID(deviceID))
			if err != nil {
				if err = checkShare(context, accounts, user); err != nil {
					return false, err
				}

				device, err = devices.Register(ctx, user, service.DeviceID(deviceID))

				if err != nil {
					return false, fmt.Errorf("device registration error: %w", err)
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

// checkShare verifies a share code.
// if there is no share, it errors out
// if there is a share code, but it does not apply for the given user account in accounts, then it errors out
// if the IsShared check succeeds, it returns no error.
func checkShare(ctx echo.Context, accounts service.Accounts, user service.Account) error {
	var err error

	share := ctx.QueryParam(ShareQueryParamName)
	if share, err = url.QueryUnescape(share); err != nil {
		return fmt.Errorf("error while unescaping share code: %w", err)
	}

	if share == "" {
		// No share Code supplied
		return echo.ErrForbidden
	}

	if shared, err := accounts.IsShared(ctx.Request().Context(), user.Username(), share); !shared || err != nil {
		// Share code provided but not valid
		return echo.ErrForbidden
	}

	return nil
}
