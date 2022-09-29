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
	ErrDeviceNotRegistered      = errors.New("device not found in account and there was no share code")
	ErrAccountShareCodeMismatch = errors.New("the provided share code did not belong to the provided account")
)

const (
	passLength, minSpecial, minNum = 32, 6, 6
)

//nolint:funlen
func (api *API) Register(ctx echo.Context, params REST.RegisterParams) error {
	var account service.Account
	var device service.Device
	var shareCode service.ShareCode

	deviceID := service.DeviceID(params.XDeviceID)
	username, password, err := basic.CredentialsFromAuthorizationHeader(ctx)

	if err != nil && err != basic.ErrNoCredentialsInHeader {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid basic auth header cannot be used for registration",
		).SetInternal(err)
	}

	// if the share code exists we have to verify it
	if params.Share != nil {
		shareCode = service.ShareCode(*params.Share)
		if account, err = api.resolveShareCode(ctx, shareCode); err != nil {
			return echo.NewHTTPError(http.StatusForbidden).SetInternal(err)
		}

		if username != "" && account.Username() != username {
			return echo.NewHTTPError(http.StatusForbidden).SetInternal(ErrAccountShareCodeMismatch)
		}

		username = account.Username()
	} else {
		account, _ = api.Accounts.Find(ctx.Request().Context(), username)
	}

	if username, err = api.defaultUsername(username); err != nil {
		return err
	}

	if password, err = api.defaultPassword(password); err != nil {
		return err
	}

	if account == nil {
		// if the account did not exist we can create it
		account, err = api.Accounts.Create(ctx.Request().Context(), username)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError).
				SetInternal(fmt.Errorf("error while creating account with provided credentials: %w", err))
		}
	} else {
		device, _ = api.Devices.GetDevice(ctx.Request().Context(), account, deviceID)

		// the device is not in the account and there is no valid share code
		if device == nil && shareCode == "" {
			return echo.NewHTTPError(http.StatusForbidden).
				SetInternal(ErrDeviceNotRegistered)
		}
	}

	// if the device is present or there is a valid shareCode is then we are free to (re-)register the device
	device, err = api.Devices.AddDevice(ctx.Request().Context(), account, deviceID, password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(
			fmt.Errorf("cannot register device %s for %s: %w", deviceID, account.Username(), err),
		)
	}

	if err = api.revokeShareCode(ctx.Request().Context(), shareCode); err != nil {
		return err
	}

	ctx.Response().Header().Set(basic.DeviceIDHeader, device.ID().String())

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

func (api *API) resolveShareCode(ctx echo.Context, share service.ShareCode) (service.Account, error) {
	// check that if the device code is present, it is actually for the account
	account, err := api.Sharing.Shared(ctx.Request().Context(), share)

	if err == service.ErrShareCodeInvalid {
		return nil, fmt.Errorf("share %s is invalid (not shared): %w", share, err)
	}

	if err != nil {
		return nil, fmt.Errorf("cannot verify share %s is valid: %w", share, err)
	}

	return account, nil
}

func (api *API) defaultUsername(username string) (string, error) {
	var err error
	if username == "" {
		// if no username is present through Basic header or the Share Code, generate it
		username, err = api.UsernameGenerator.Generate()
		if err != nil {
			return "", echo.NewHTTPError(http.StatusInternalServerError).SetInternal(
				fmt.Errorf("generating a username for registration failed: %w", err),
			)
		}
	}

	return username, err
}

func (api *API) defaultPassword(password string) (string, error) {
	var err error
	if password == "" {
		// if no password is present through Basic header, generate it
		password, err = api.PasswordGenerator.Generate(passLength, minNum, minSpecial, false, false)
		if err != nil {
			return "", echo.NewHTTPError(http.StatusInternalServerError).SetInternal(
				fmt.Errorf("generating a password for registration failed: %w", err),
			)
		}
	}

	return password, err
}

func (api *API) revokeShareCode(ctx context.Context, code service.ShareCode) error {
	var err error
	if code != "" {
		if err = api.Sharing.Revoke(ctx, code); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(
				fmt.Errorf("cannot revoke old share code %s: %w", code, err),
			)
		}
	}

	return err
}
