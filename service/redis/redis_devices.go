package redis

import (
	"context"
	"fmt"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"

	"github.com/go-redis/redis/v9"
)

const DeviceKeySpace = "octi:devices"

func NewDevices(client *redis.Client) *Devices {
	return &Devices{client}
}

type Devices struct {
	client *redis.Client
}

func (r *Devices) deviceKeyForAccount(acc service.Account) string {
	return fmt.Sprintf("%s:%s", DeviceKeySpace, acc.Username())
}

func (r *Devices) FindByAccount(ctx context.Context, acc service.Account) ([]service.Device, error) {
	res, err := r.client.LRange(ctx, r.deviceKeyForAccount(acc), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	devices := make([]service.Device, len(res))
	for i, deviceID := range res {
		devices[i] = DeviceFromID(deviceID)
	}
	return devices, nil
}

func (r *Devices) FindByDeviceID(ctx context.Context, acc service.Account, deviceID string) (service.Device, error) {
	res, err := r.client.LPos(ctx, r.deviceKeyForAccount(acc), deviceID, redis.LPosArgs{}).Result()
	if err != nil {
		return nil, err
	}
	if res >= 0 {
		return DeviceFromID(deviceID), nil
	}
	return nil, service.ErrDeviceNotFound
}

func (r *Devices) Register(ctx context.Context, acc service.Account, deviceID string) error {
	return r.client.LPush(ctx, r.deviceKeyForAccount(acc), deviceID).Err()
}

func (r *Devices) HealthCheck() service.HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "redis-devices", r.client.Ping(ctx).Err() == nil
	}
}
