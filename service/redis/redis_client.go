package redis

import (
	"context"
	"time"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	goredis "github.com/redis/go-redis/v9"
)

const (
	EnvRedisAddr           = "REDIS_ADDR"
	EnvRedisUsername       = "REDIS_USERNAME"
	EnvRedisPassword       = "REDIS_PASSWORD"
	DefaultIntervalSeconds = 5
	DefaultTimeoutSeconds  = 5
	NoExpiry               = time.Duration(-1)
)

type (
	Clients        map[string]goredis.Cmdable
	ClientMutators map[string]ClientMutator
	ClientMutator  func(client goredis.UniversalClient) goredis.UniversalClient
)

type ClientProvider func(config *config.Config) goredis.UniversalClient

//go:generate mockgen -package mock -destination mock/redis.go github.com/redis/go-redis/v9 UniversalClient
func NewClientsWithRegularPing(
	ctx context.Context,
	config *config.Config,
	provider ClientProvider,
	mutators ClientMutators,
) (Clients, error) {
	logger := config.Logger
	applyDefaultConfiguration(logger, config)

	client := provider(config)

	detailLogger := logger.With().Logger()

	if config.Redis.Username != "" {
		detailLogger = logger.With().Str("username", config.Redis.Username).Logger()
	}

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

	if len(mutators) == 0 {
		clients[""] = client
	}

	return clients, nil
}
