package v1

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/redis"
	"github.com/labstack/echo/v4"
)

const XModifiedAt = "X-Modified-At"

var ErrAccountForVerifyingDeviceNotPresent = errors.New("account for verifying device id is not present")

func (api *API) CreateModule(ctx echo.Context, name REST.ModuleName, params REST.CreateModuleParams) error {
	acc, accountPresent := ctx.Get(basic.AccountKey).(service.Account)

	if !accountPresent {
		return echo.NewHTTPError(
			http.StatusForbidden, ErrAccountForVerifyingDeviceNotPresent,
		)
	}

	id := fmt.Sprintf("%s-%s-%s", acc.Username(), params.XDeviceID, name)

	modifiedAt := time.Now()

	err := api.Modules.Set(
		ctx.Request().Context(),
		id,
		redis.ModuleFromReader(ctx.Request().Body, int(ctx.Request().ContentLength)),
	)
	if err != nil {
		return fmt.Errorf("could not create/update module: %w", err)
	}

	err = api.MetadataProvider.Set(
		ctx.Request().Context(), service.NewBaseMetadata(id, modifiedAt),
	)
	if err != nil {
		return fmt.Errorf("could not create/update module metadata: %w", err)
	}

	if err := ctx.JSON(http.StatusAccepted, nil); err != nil {
		return fmt.Errorf("could not acknowledge module creation: %w", err)
	}

	return nil
}

func (api *API) GetModule(ctx echo.Context, name REST.ModuleName, params REST.GetModuleParams) error {
	deviceID := params.XDeviceID

	acc, accountPresent := ctx.Get(basic.AccountKey).(service.Account)

	if !accountPresent {
		return echo.NewHTTPError(
			http.StatusForbidden, ErrAccountForVerifyingDeviceNotPresent,
		)
	}

	if params.DeviceId != nil {
		deviceID = *params.DeviceId
		if _, err := api.GetDevice(ctx.Request().Context(), acc, service.DeviceID(deviceID)); err != nil {
			return echo.NewHTTPError(
				http.StatusForbidden, fmt.Errorf(
					"device from params could not be verified against account: %w", err,
				),
			)
		}
	}

	id := fmt.Sprintf("%s-%s-%s", acc.Username(), deviceID, name)

	module, err := api.Modules.Get(ctx.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("error while fetching module: %w", err)
	}

	metadata, err := api.MetadataProvider.Get(
		ctx.Request().Context(), service.MetadataID(id),
	)
	if err != nil {
		return fmt.Errorf("could not create/update module metadata: %w", err)
	}

	ctx.Response().Header().Set(
		XModifiedAt,
		REST.ModifiedAtTimestamp(metadata.GetModifiedAt()).Format(time.RFC3339),
	)

	if err := ctx.Stream(http.StatusOK, echo.MIMEOctetStream, module.Raw()); err != nil {
		return fmt.Errorf("error while writing module data to response while fetching module: %w", err)
	}

	return nil
}
