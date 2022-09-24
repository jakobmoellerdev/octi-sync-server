package redis

import (
	"context"

	goredis "github.com/go-redis/redis/v9"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
)

const (
	EnvRedisAddr           = "REDIS_ADDR"
	EnvRedisUsername       = "REDIS_USERNAME"
	EnvRedisPassword       = "REDIS_PASSWORD"
	DefaultIntervalSeconds = 5
	DefaultTimeoutSeconds  = 5
)

type (
	Clients        map[string]goredis.Cmdable
	ClientMutators map[string]ClientMutator
	ClientMutator  func(client goredis.UniversalClient) goredis.UniversalClient
)

//go:generate mockgen -package mock -destination mock/redis.go github.com/go-redis/redis/v9 UniversalClient
func NewClientsWithRegularPing(ctx context.Context, config *config.Config, mutators ClientMutators) (Clients, error) {
	logger := config.Logger
	applyDefaultConfiguration(logger, config)

	client := goredis.NewUniversalClient(&config.Redis.UniversalOptions)

	detailLogger := logger.With().
		Str("username", config.Redis.Username).
		Logger()

	logger = &detailLogger

	if config.Redis.Ping.Enable {
		StartPingingRedis(ctx, config.Redis.Ping.Interval, client, logger)
	}

	clients := Clients{}

	for mutatorName, mutator := range mutators {
		if mutator == nil {
			clients[mutatorName] = client
		} else {
			clients[mutatorName] = mutator(client)
		}
	}

	return clients, nil
}
