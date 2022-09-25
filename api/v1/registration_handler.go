package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
)

var ErrPasswordMismatch = errors.New("passwords do not match")

func (api *API) Register(ctx echo.Context, params REST.RegisterParams) error {
	var err error
	var username, password string
	var account service.Account

	deviceID := service.DeviceID(params.XDeviceID)

	if username, password, err = basic.CredentialsFromAuthorizationHeader(ctx); err == echo.ErrBadRequest {
		return err
	}

	if err == basic.ErrNoCredentialsInHeader {
		// if no credentials are present through Basic header, generate username and password
		if username, password, err = api.newCredentials(); err != nil {
			return err
		}

		if account, err = api.Accounts.Register(ctx.Request().Context(), username, password); err != nil {
			return fmt.Errorf("account could not be registered (username: %s): %w", username, err)
		}
	} else {
		// for present credentials, verify the Basic header
		if account, err = api.Accounts.Find(ctx.Request().Context(), username); err != nil {
			return echo.NewHTTPError(http.StatusForbidden).SetInternal(err)
		} else if !account.Verify(password) {
			return echo.NewHTTPError(http.StatusForbidden).SetInternal(ErrPasswordMismatch)
		}
	}

	// next use the device-id from the parameters
	if _, err = api.Devices.FindByDeviceID(ctx.Request().Context(), account, deviceID); err == service.ErrDeviceNotFound {
		// if the device does not exist we have to verify the share code
		if params.Share == nil {
			return echo.NewHTTPError(http.StatusForbidden).SetInternal(err)
		}

		if err = api.verifyShareCode(ctx, account, *params.Share); err != nil {
			return echo.NewHTTPError(http.StatusForbidden).SetInternal(err)
		}

		// if it is then we are free to register the device
		if _, err := api.Devices.Register(ctx.Request().Context(), account, deviceID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(
				fmt.Errorf("cannot register device %s for %s: %w", deviceID, account.Username(), err))
		}
	}

	ctx.Response().Header().Set(basic.DeviceIDHeader, deviceID.String())

	if err := ctx.JSON(http.StatusOK, &REST.RegistrationResult{
		Password: password,
		Username: account.Username(),
	}); err != nil {
		return fmt.Errorf("could not write registration response: %w", err)
	}

	return nil
}

func (api *API) newCredentials() (string, string, error) {
	var passLength, minSpecial, minNum = 32, 6, 6
	var username, password string

	username, err := api.UsernameGenerator.Generate()

	if err != nil {
		return "", "", fmt.Errorf("generating a username for registration failed: %w", err)
	}

	password, err = api.PasswordGenerator.Generate(passLength, minNum, minSpecial, false, false)

	if err != nil {
		return "", "", fmt.Errorf("generating a password for registration failed: %w", err)
	}

	return username, password, nil
}

func (api *API) verifyShareCode(ctx echo.Context, account service.Account, share REST.ShareCode) error {
	// check that if the device code is present, it is actually for the account
	if err := api.Accounts.IsShared(ctx.Request().Context(), account.Username(), share); err == service.ErrShareCodeInvalid {
		return fmt.Errorf("share %s is invalid (not shared) for %s: %w", share, account.Username(), err)
	} else if err != nil {
		return fmt.Errorf("cannot verify share %s is valid for %s: %w", share, account.Username(), err)
	}

	return nil
}
