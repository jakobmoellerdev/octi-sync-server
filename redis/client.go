package redis

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"

	"github.com/go-redis/redis/v9"
)

var ErrNoAddrProvided = errors.New("no redis address provided")

const (
	EnvRedisAddr           = "REDIS_ADDR"
	EnvRedisUsername       = "REDIS_USERNAME"
	EnvRedisPassword       = "REDIS_PASSWORD"
	DefaultIntervalSeconds = 5
	DefaultTimeoutSeconds  = 5
)

func NewClientWithRegularPing(ctx context.Context, config *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&config.Redis.Options)

	if err := applyDefaultConfiguration(config); err != nil {
		return nil, err
	}

	if config.Redis.Ping.Interval <= 0 {
		config.Redis.Ping.Interval = DefaultIntervalSeconds * time.Second
		config.Logger.Info("defaulting redis ping interval to " + config.Redis.Ping.Interval.String())
	}

	if config.Redis.Ping.Timeout <= 0 {
		config.Redis.Ping.Timeout = DefaultTimeoutSeconds * time.Second
		config.Logger.Info("defaulting redis ping timeout to " + config.Redis.Ping.Timeout.String())
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

func applyDefaultConfiguration(config *config.Config) error {
	if config.Redis.Addr == "" {
		config.Logger.Info("config does not contain address, defaulting redis addr to " + EnvRedisAddr)
		addrFromEnv, found := os.LookupEnv(EnvRedisAddr)
		if !found {
			return ErrNoAddrProvided
		}
		config.Redis.Addr = addrFromEnv
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
	return nil
}

func VerifyConnection(ctx context.Context, client *redis.Client, timeout time.Duration) error {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	_, err := client.Ping(pingCtx).Result()
	return err
}
