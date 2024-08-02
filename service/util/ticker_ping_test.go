package util_test

import (
	"context"
	"testing"
	"time"

	"github.com/jakobmoellerdev/octi-sync-server/service/util"
)

func TestTickerPinger(t *testing.T) {
	t.Parallel()

	pinger := util.NewIntervalTickerPinger(1*time.Microsecond, func(ctx context.Context) {})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Microsecond)
	defer cancel()

	go pinger.Start(ctx)

	pinger.Stop()
}
