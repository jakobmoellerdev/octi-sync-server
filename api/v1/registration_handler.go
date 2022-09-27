package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/labstack/echo/v4"
)

var (
	ErrPasswordMismatch    = errors.New("passwords do not match")
	ErrDeviceNotRegistered = errors.New("device not found in account and there was no share code")
)

//nolint:funlen
func (api *API) Register(ctx echo.Context, params REST.RegisterParams) error {
	deviceID := service.DeviceID(params.XDeviceID)
	username, password, err := basic.CredentialsFromAuthorizationHeader(ctx)

	if err != nil && err != basic.ErrNoCredentialsInHeader {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid basic auth header cannot be used for registration",
		).SetInternal(err)
	}

	var account service.Account
	var device service.Device

	if err == basic.ErrNoCredentialsInHeader {
		account, device, err = api.newAccount(ctx.Request().Context(), deviceID)
	} else {
		account, device, err = api.newOrExistingAccount(ctx.Request().Context(), deviceID, username, password)
	}

	if err != nil {
		return err
	}

	// next use the device-id from the parameters
	if device == nil {
		// if the device does not exist we have to verify the share code
		if params.Share == nil {
			return echo.NewHTTPError(http.StatusForbidden).SetInternal(ErrDeviceNotRegistered)
		}

		shareCode := service.ShareCode(*params.Share)

		if err = api.verifyShareCode(ctx, account, shareCode); err != nil {
			return echo.NewHTTPError(http.StatusForbidden).SetInternal(err)
		}

		// if it is then we are free to register the device
		if _, err = api.Devices.AddDevice(ctx.Request().Context(), account, deviceID, password); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(
				fmt.Errorf("cannot register device %s for %s: %w", deviceID, account.Username(), err),
			)
		}

		if err = api.Sharing.Revoke(ctx.Request().Context(), account, shareCode); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(
				fmt.Errorf("cannot revoke old share code %s for %s: %w", deviceID, account.Username(), err),
			)
		}
	}

	ctx.Response().Header().Set(basic.DeviceIDHeader, deviceID.String())

	if err = ctx.JSON(
		http.StatusOK, &REST.RegistrationResult{
			Password: password,
			Username: account.Username(),
		},
	); err != nil {
		return fmt.Errorf("could not write registration response: %w", err)
	}

	return nil
}

func (api *API) newOrExistingAccount(
	ctx context.Context,
	deviceID service.DeviceID,
	username, password string,
) (service.Account, service.Device, error) {
	account, err := api.Accounts.Find(ctx, username)

	if err == service.ErrAccountNotFound {
		account, err = api.Accounts.Create(ctx, username)
		if err != nil {
			return nil, nil, echo.NewHTTPError(http.StatusInternalServerError).
				SetInternal(fmt.Errorf("error while creating account with provided credentials: %w", err))
		}
		device, err := api.Devices.AddDevice(ctx, account, deviceID, password)
		if err != nil {
			return nil, nil, echo.NewHTTPError(http.StatusInternalServerError).
				SetInternal(fmt.Errorf("registering a new device failed: %w", err))
		}
		return account, device, nil
	}

	if err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("error while finding account to verify credentials: %w", err))
	}

	device, _ := api.Devices.GetDevice(ctx, account, deviceID)

	if device != nil && !device.Verify(password) {
		return nil, nil, echo.NewHTTPError(http.StatusForbidden).SetInternal(ErrPasswordMismatch)
	}

	return account, device, nil
}

func (api *API) newAccount(
	ctx context.Context,
	deviceID service.DeviceID,
) (service.Account, service.Device, error) {
	// if no credentials are present through Basic header, generate username and password

	username, err := api.UsernameGenerator.Generate()
	if err != nil {
		return nil, nil, fmt.Errorf("generating a username for registration failed: %w", err)
	}

	passLength, minSpecial, minNum := 32, 6, 6

	password, err := api.PasswordGenerator.Generate(passLength, minNum, minSpecial, false, false)
	if err != nil {
		return nil, nil, fmt.Errorf("generating a password for registration failed: %w", err)
	}

	account, err := api.Accounts.Create(ctx, username)
	if err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("account could not be registered (username: %s): %w", username, err))
	}

	device, err := api.Devices.AddDevice(ctx, account, deviceID, password)
	if err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("error while registering new device: %w", err))
	}

	return account, device, nil
}

func (api *API) verifyShareCode(ctx echo.Context, account service.Account, share service.ShareCode) error {
	// check that if the device code is present, it is actually for the account
	err := api.Sharing.IsShared(ctx.Request().Context(), account, share)

	if err == service.ErrShareCodeInvalid {
		return fmt.Errorf("share %s is invalid (not shared) for %s: %w", share, account.Username(), err)
	}

	if err != nil {
		return fmt.Errorf("cannot verify share %s is valid for %s: %w", share, account.Username(), err)
	}

	return nil
}
