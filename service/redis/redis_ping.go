package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/jakob-moeller-cloud/octi-sync-server/service/util"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func StartPingingRedis(
	ctx context.Context,
	interval time.Duration,
	client redis.Cmdable,
	logger *zerolog.Logger,
) {
	ping := func(ctx context.Context) {
		if err := VerifyConnection(ctx, client, interval); err != nil {
			logger.Warn().Msg("redis client could not verify connection, ping failed")
		} else {
			logger.Debug().Msg("redis ping ok!")
		}
	}

	go util.NewIntervalTickerPinger(interval, ping).Start(ctx)
}

func VerifyConnection(ctx context.Context, client redis.Cmdable, timeout time.Duration) error {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := client.Ping(pingCtx).Result()
	if err != nil {
		return fmt.Errorf("error while verifying connection with ping: %w", err)
	}

	return nil
}
