package redis

import (
	"context"
	"crypto/sha256"
	"fmt"

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
	ClientMutator  func(client *goredis.Client) *goredis.Client
)

//go:generate mockgen -package mock -destination mock/redis.go github.com/go-redis/redis/v9 Cmdable
func NewClientsWithRegularPing(ctx context.Context, config *config.Config, mutators ClientMutators) (Clients, error) {
	client := goredis.NewClient(&config.Redis.Options)
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
