package redis

import (
	"os"
	"time"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/rs/zerolog"
)

func applyDefaultConfiguration(logger *zerolog.Logger, config *config.Config) {
	if config.Redis.Ping.Interval <= 0 {
		config.Redis.Ping.Interval = DefaultIntervalSeconds * time.Second
		logger.Info().Msg("defaulting redis ping interval to " + config.Redis.Ping.Interval.String())
	}

	if config.Redis.Ping.Timeout <= 0 {
		config.Redis.Ping.Timeout = DefaultTimeoutSeconds * time.Second
		logger.Info().Msg("defaulting redis ping timeout to " + config.Redis.Ping.Timeout.String())
	}

	if config.Redis.Addrs == nil || config.Redis.Addrs[0] == "localhost:6379" {
		if addrFromEnv, found := os.LookupEnv(EnvRedisAddr); found {
			logger.Info().Msg("config does not contain address, defaulting redis addr to " + EnvRedisAddr)

			config.Redis.Addrs = []string{addrFromEnv}
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
