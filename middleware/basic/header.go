package basic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	Basic       = "Basic"
	lengthBasic = len(Basic)
)

var ErrNoCredentialsInHeader = errors.New("no basic auth credentials in header")

func CredentialsFromAuthorizationHeader(ctx echo.Context) (string, string, error) {
	auth := ctx.Request().Header.Get(echo.HeaderAuthorization)

	if len(auth) > lengthBasic+1 && strings.EqualFold(auth[:lengthBasic], Basic) {
		// Invalid base64 shouldn't be treated as error
		// instead should be treated as invalid client input
		b, err := base64.StdEncoding.DecodeString(auth[lengthBasic+1:])
		cred := strings.Split(string(b), ":")

		if err != nil || len(cred) != 2 {
			return "", "", fmt.Errorf("basic auth credentials in header invalid: %w", err)
		}

		// use user and pass from Basic header if present
		return cred[0], cred[1], nil
	}

	return "", "", ErrNoCredentialsInHeader
}
