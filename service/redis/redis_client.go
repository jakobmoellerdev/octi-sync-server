package redis

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/go-redis/redis/v9"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
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
		StartPingingRedis(ctx, config.Redis.Ping.Interval, client, logger)
	}

	return client, nil
}
