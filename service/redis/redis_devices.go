package redis

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/redis/go-redis/v9"
)

const DeviceKeySpace = "octi:devices"

type Devices struct {
	Client redis.Cmdable
}

func (r *Devices) deviceKeyForAccount(acc service.Account) string {
	return fmt.Sprintf("%s:%s", DeviceKeySpace, acc.Username())
}

func (r *Devices) hashPassword(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}

func (r *Devices) AddDevice(
	ctx context.Context, account service.Account, id service.DeviceID, password string,
) (service.Device, error) {
	hashed := r.hashPassword(password)

	if err := r.Client.HSet(ctx, r.deviceKeyForAccount(account), id.String(), hashed).Err(); err != nil {
		return nil, fmt.Errorf("could not push device id for registration: %w", err)
	}

	return service.NewBaseDevice(id, hashed), nil
}

func (r *Devices) GetDevices(
	ctx context.Context,
	account service.Account,
) (map[service.DeviceID]service.Device, error) {
	res, err := r.Client.HGetAll(ctx, r.deviceKeyForAccount(account)).Result()
	if err != nil {
		return nil, fmt.Errorf("could not find devices by account: %w", err)
	}

	devices := make(map[service.DeviceID]service.Device, len(res))

	for id, pass := range res {
		deviceUUID, err := uuid.Parse(id)
		if err != nil {
			return devices, fmt.Errorf("device id could not be parsed: %w", err)
		}

		deviceID := service.DeviceID(deviceUUID)
		devices[deviceID] = service.NewBaseDevice(deviceID, pass)
	}

	return devices, nil
}

func (r *Devices) GetDevice(ctx context.Context, account service.Account, id service.DeviceID) (service.Device, error) {
	res, err := r.Client.HGet(ctx, r.deviceKeyForAccount(account), id.String()).Result()

	if err == redis.Nil {
		return nil, service.ErrDeviceNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("could not find devices by id: %w", err)
	}

	return service.NewBaseDevice(id, res), nil
}

func (r *Devices) DeleteDevice(
	ctx context.Context,
	account service.Account,
	id service.DeviceID,
) error {
	if err := r.Client.HDel(ctx, r.deviceKeyForAccount(account), id.String()).Err(); err != nil {
		return fmt.Errorf("deletion of device failed: %w", err)
	}

	return nil
}

func (r *Devices) HealthCheck() service.HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "redis-devices", r.Client.Ping(ctx).Err() == nil
	}
}
