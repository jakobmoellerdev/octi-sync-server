package service

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v9"
)

const RedisDeviceKeySpace = "octi:devices"

func NewRedisDevices(client *redis.Client) *RedisDevices {
	return &RedisDevices{client}
}

type RedisDevices struct {
	client *redis.Client
}

func (r *RedisDevices) deviceKeyForAccount(acc Account) string {
	return fmt.Sprintf("%s:%s", RedisDeviceKeySpace, acc.Username())
}

func (r *RedisDevices) FindByAccount(ctx context.Context, acc Account) ([]Device, error) {
	res, err := r.client.LRange(ctx, r.deviceKeyForAccount(acc), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	devices := make([]Device, len(res))
	for i, deviceID := range res {
		devices[i] = RedisDeviceFromID(deviceID)
	}
	return devices, nil
}

func (r *RedisDevices) FindByDeviceID(ctx context.Context, acc Account, deviceID string) (Device, error) {
	res, err := r.client.LPos(ctx, r.deviceKeyForAccount(acc), deviceID, redis.LPosArgs{}).Result()
	if err != nil {
		return nil, err
	}
	if res >= 0 {
		return RedisDeviceFromID(deviceID), nil
	}
	return nil, ErrDeviceNotFound
}

func (r *RedisDevices) Register(ctx context.Context, acc Account, deviceID string) error {
	return r.client.LPush(ctx, r.deviceKeyForAccount(acc), deviceID).Err()
}

func (r *RedisDevices) HealthCheck() HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "redis-devices", r.client.Ping(ctx).Err() == nil
	}
}
