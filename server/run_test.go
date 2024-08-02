package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/jakobmoellerdev/octi-sync-server/config"
	"github.com/jakobmoellerdev/octi-sync-server/server"
)

func TestRun(t *testing.T) {
	log := zerolog.New(zerolog.NewTestWriter(t))

	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context, t *testing.T) {
		assert.New(t).NoError(
			server.Run(
				ctx, &config.Config{
					Logger: &log,
				},
			),
		)
	}(ctx, t)
	time.Sleep(100 * time.Millisecond)
	cancel()
}
