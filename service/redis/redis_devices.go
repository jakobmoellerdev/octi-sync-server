package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

const DeviceKeySpace = "octi:devices"

type Devices struct {
	Client redis.Cmdable
}

func (r *Devices) deviceKeyForAccount(acc service.Account) string {
	return fmt.Sprintf("%s:%s", DeviceKeySpace, acc.Username())
}

func (r *Devices) FindByAccount(ctx context.Context, acc service.Account) ([]service.Device, error) {
	res, err := r.Client.LRange(ctx, r.deviceKeyForAccount(acc), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("could not find devices by account: %w", err)
	}

	devices := make([]service.Device, len(res))

	for i, id := range res {
		deviceId, err := uuid.Parse(id)
		if err != nil {
			return devices, fmt.Errorf("device id could not be parsed: %w", err)
		}

		devices[i] = service.NewBaseDevice(service.DeviceID(deviceId))
	}

	return devices, nil
}

func (r *Devices) FindByDeviceID(
	ctx context.Context,
	acc service.Account,
	deviceID service.DeviceID,
) (service.Device, error) {
	res, err := r.Client.LPos(ctx, r.deviceKeyForAccount(acc), deviceID.String(), redis.LPosArgs{}).Result()

	if err == redis.Nil || res < 0 {
		return nil, service.ErrDeviceNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("could not find devices by id: %w", err)
	}

	return service.NewBaseDevice(deviceID), nil
}

func (r *Devices) Register(
	ctx context.Context,
	acc service.Account,
	deviceID service.DeviceID,
) (service.Device, error) {
	if err := r.Client.LPush(ctx, r.deviceKeyForAccount(acc), deviceID.String()).Err(); err != nil {
		return nil, fmt.Errorf("could not push device id for registration: %w", err)
	}

	return service.NewBaseDevice(deviceID), nil
}

func (r *Devices) HealthCheck() service.HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "redis-devices", r.Client.Ping(ctx).Err() == nil
	}
}
