package util

import (
	"context"
	"time"
)

type TickerPinger interface {
	Start(ctx context.Context)
	Stop()
}

type OnPing func(ctx context.Context)

func NewIntervalTickerPinger(
	interval time.Duration,
	OnPing OnPing,
) TickerPinger {
	return &tickerPinger{time.NewTicker(interval), interval, OnPing}
}

type tickerPinger struct {
	ticker   *time.Ticker
	interval time.Duration
	OnPing
}

func (p *tickerPinger) Start(ctx context.Context) {
	pingCtx, pingCancel := context.WithCancel(ctx)
	defer pingCancel()

	for {
		select {
		case <-pingCtx.Done():
			p.ticker.Stop()

			return
		case <-p.ticker.C:
			p.OnPing(pingCtx)
		}
	}
}

func (p *tickerPinger) Stop() {
	p.ticker.Stop()
}
