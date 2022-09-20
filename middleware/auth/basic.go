package auth

import (
	"crypto/subtle"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// UserKey is the cookie name for user credential in basic auth.
const UserKey = "user"

// BasicAuth returns a Basic HTTP Authorization middleware. It takes as argument a map[string]string where
// the key is the username and the value is the password.
func BasicAuth(accounts service.Accounts) echo.MiddlewareFunc {
	return middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper: middleware.DefaultSkipper,
		Validator: func(username, password string, context echo.Context) (bool, error) {
			// Search user in the slice of allowed credentials
			user, err := accounts.Find(context.Request().Context(), username)
			if err != nil {
				return false, nil
			}

			if subtle.ConstantTimeCompare([]byte(user.HashedPass()), []byte(password)) != 1 {
				return false, nil
			}
			// The user credentials was found, set user's id to key DeviceID in this context,
			// the user's id can be read later using
			// context.MustGet(auth.DeviceID).
			context.Set(UserKey, user)
			return true, nil
		},
		Realm: "",
	})
}
