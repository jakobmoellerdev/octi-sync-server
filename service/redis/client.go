package redis

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/rs/zerolog"
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
	logger := config.Logger

	applyDefaultConfiguration(logger, config)
	detailLogger := logger.With().
		Str("client", client.String()).
		Str("username", config.Redis.Username).
		Str("pass", fmt.Sprintf("%x", sha256.Sum256([]byte(config.Redis.Password)))).
		Logger()
	logger = &detailLogger

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
						logger.Warn().Msg("redis client could not verify connection, ping failed")
					} else {
						logger.Debug().Msg("redis ping ok!")
					}
				}
			}
		}

		go verify()
	}

	return client, nil
}

func applyDefaultConfiguration(logger *zerolog.Logger, config *config.Config) {
	if config.Redis.Ping.Interval <= 0 {
		config.Redis.Ping.Interval = DefaultIntervalSeconds * time.Second
		logger.Info().Msg("defaulting redis ping interval to " + config.Redis.Ping.Interval.String())
	}

	if config.Redis.Ping.Timeout <= 0 {
		config.Redis.Ping.Timeout = DefaultTimeoutSeconds * time.Second
		logger.Info().Msg("defaulting redis ping timeout to " + config.Redis.Ping.Timeout.String())
	}

	if config.Redis.Addr == "localhost:6379" {
		if addrFromEnv, found := os.LookupEnv(EnvRedisAddr); found {
			logger.Info().Msg("config does not contain address, defaulting redis addr to " + EnvRedisAddr)

			config.Redis.Addr = addrFromEnv
		} else {
			logger.Info().Msg("connecting against localhost instead of connecting remotely")
		}
	}

	if config.Redis.Username == "" {
		if usernameFromEnv, found := os.LookupEnv(EnvRedisUsername); found {
			logger.Info().Msg("config does not contain username, defaulting to " + EnvRedisUsername)

			config.Redis.Username = usernameFromEnv
		}
	}

	if config.Redis.Password == "" {
		if passwordFromEnv, found := os.LookupEnv(EnvRedisPassword); found {
			logger.Info().Msg("config does not contain password, defaulting to " + EnvRedisPassword)

			config.Redis.Password = passwordFromEnv
		}
	}
}

func VerifyConnection(ctx context.Context, client *redis.Client, timeout time.Duration) error {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := client.Ping(pingCtx).Result()

	return fmt.Errorf("error while verifying connection with ping: %w", err)
}
