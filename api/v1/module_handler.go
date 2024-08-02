package v1

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/jakobmoellerdev/octi-sync-server/api/v1/REST"
	"github.com/jakobmoellerdev/octi-sync-server/middleware/basic"
	"github.com/jakobmoellerdev/octi-sync-server/service"
	"github.com/jakobmoellerdev/octi-sync-server/service/redis"
)

const XModifiedAt = "X-Modified-At"

var ErrAccountForVerifyingDeviceNotPresent = errors.New("account for verifying device id is not present")

func (api *API) CreateModule(ctx echo.Context, name REST.ModuleName, params REST.CreateModuleParams) error {
	acc, device, err := api.resolveDeviceIDAndAccount(ctx, &params.XDeviceID)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s-%s-%s", acc.Username(), device.ID(), name)

	modifiedAt := time.Now()

	err = api.Modules.Set(
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
	acc, device, err := api.resolveDeviceIDAndAccount(ctx, params.DeviceId, &params.XDeviceID)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s-%s-%s", acc.Username(), device.ID(), name)

	module, err := api.Modules.Get(ctx.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("error while fetching module: %w", err)
	}

	status := http.StatusOK
	if module.Size() == 0 {
		status = http.StatusNoContent
	} else {
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
	}

	if err := ctx.Stream(status, echo.MIMEOctetStream, module.Raw()); err != nil {
		return fmt.Errorf("error while writing module data to response while fetching module: %w", err)
	}

	return nil
}

func (api *API) DeleteModules(ctx echo.Context, params REST.DeleteModulesParams) error {
	acc, device, err := api.resolveDeviceIDAndAccount(ctx, params.DeviceId, &params.XDeviceID)
	if err != nil {
		return err
	}

	idPattern := fmt.Sprintf("%s-%s-*", acc.Username(), device.ID())

	if err := api.Modules.DeleteByPattern(ctx.Request().Context(), idPattern); err != nil {
		return fmt.Errorf("error while fetching module: %w", err)
	}

	if err := ctx.JSON(http.StatusAccepted, nil); err != nil {
		return fmt.Errorf("could not acknowledge module creation: %w", err)
	}

	return nil
}

func (api *API) resolveDeviceIDAndAccount(
	ctx echo.Context, deviceIDs ...*uuid.UUID,
) (service.Account, service.Device, error) {
	acc, accountPresent := ctx.Get(basic.AccountKey).(service.Account)

	if !accountPresent {
		return nil, nil, echo.NewHTTPError(
			http.StatusForbidden, ErrAccountForVerifyingDeviceNotPresent,
		)
	}

	var device service.Device

	for i := range deviceIDs {
		if deviceIDs[i] != nil {
			id := *deviceIDs[i]
			var err error

			deviceFromCtx := false
			if device, deviceFromCtx = ctx.Get(basic.Device).(service.Device); deviceFromCtx &&
				device.ID().UUID().String() == deviceIDs[i].String() {
				break
			}

			if device, err = api.GetDevice(ctx.Request().Context(), acc, service.DeviceID(id)); err != nil {
				return nil, nil, echo.NewHTTPError(
					http.StatusForbidden, fmt.Errorf(
						"device from params could not be verified against account: %w", err,
					),
				)
			}

			break
		}
	}

	return acc, device, nil
}
