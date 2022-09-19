package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.jakob-moeller.cloud/octi-sync-server/config"

	"github.com/go-redis/redis/v9"
)

const (
	EnvRedisAddr           = "REDIS_ADDR"
	EnvRedisUsername       = "REDIS_USERNAME"
	EnvRedisPassword       = "REDIS_PASSWORD"
	DefaultIntervalSeconds = 5
	DefaultTimeoutSeconds  = 5
)

func NewClientWithRegularPing(ctx context.Context, config *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&config.Redis.Options)

	applyDefaultConfiguration(config)

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
						config.Logger.Warn(fmt.Sprintf(
							"redis client could not verify connection to %s as %s, ping failed",
							config.Redis.Addr, config.Redis.Username))
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

func applyDefaultConfiguration(config *config.Config) {
	if config.Redis.Ping.Interval <= 0 {
		config.Redis.Ping.Interval = DefaultIntervalSeconds * time.Second
		config.Logger.Info("defaulting redis ping interval to " + config.Redis.Ping.Interval.String())
	}

	if config.Redis.Ping.Timeout <= 0 {
		config.Redis.Ping.Timeout = DefaultTimeoutSeconds * time.Second
		config.Logger.Info("defaulting redis ping timeout to " + config.Redis.Ping.Timeout.String())
	}

	if config.Redis.Addr == "localhost:6379" {
		addrFromEnv, found := os.LookupEnv(EnvRedisAddr)
		if found {
			config.Logger.Info("config does not contain address, defaulting redis addr to " + EnvRedisAddr)
			config.Redis.Addr = addrFromEnv
		} else {
			config.Logger.Info("connecting against localhost instead of connecting remotely")
		}
	}
	if config.Redis.Username == "" {
		usernameFromEnv, found := os.LookupEnv(EnvRedisUsername)
		if found {
			config.Logger.Info("config does not contain username, defaulting to " + EnvRedisUsername)
			config.Redis.Username = usernameFromEnv
		}
	}
	if config.Redis.Password == "" {
		passwordFromEnv, found := os.LookupEnv(EnvRedisPassword)
		if found {
			config.Logger.Info("config does not contain password, defaulting to " + EnvRedisPassword)
			config.Redis.Password = passwordFromEnv
		}
	}
}

func VerifyConnection(ctx context.Context, client *redis.Client, timeout time.Duration) error {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	_, err := client.Ping(pingCtx).Result()
	return err
}
