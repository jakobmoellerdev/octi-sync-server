package redis

import (
	"context"
	"github.com/go-redis/redis/v9"
	"octi-sync-server/config"
	"time"
)

var DefaultRedisPingInterval = 5 * time.Second
var DefaultRedisPingTimeout = 5 * time.Second

func NewClientWithRegularPing(ctx context.Context, config *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&config.Redis.Options)

	if config.Redis.Ping.Interval <= 0 {
		config.Redis.Ping.Interval = DefaultRedisPingInterval
		config.Logger.Info("defaulting redis ping interval to " + DefaultRedisPingInterval.String())
	}

	if config.Redis.Ping.Timeout <= 0 {
		config.Redis.Ping.Timeout = DefaultRedisPingTimeout
		config.Logger.Info("defaulting redis ping timeout to " + DefaultRedisPingTimeout.String())
	}

	if config.Redis.Ping.Enable {
		ticker := time.NewTicker(config.Redis.Ping.Interval)
		verify := func() {
			for {
				select {
				case <-ctx.Done():
					ticker.Stop()
					return
				case <-ticker.C:
					if err := VerifyConnection(ctx, client, config.Redis.Ping.Timeout); err != nil {
						config.Logger.Warn("redis client could not verify connection, ping failed")
					} else {
						config.Logger.Debug("redis ping ok!")
					}
				}
			}

		}
		go verify()
	}

	return client, nil
}

func VerifyConnection(ctx context.Context, client *redis.Client, timeout time.Duration) error {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	_, err := client.Ping(pingCtx).Result()
	return err
}
