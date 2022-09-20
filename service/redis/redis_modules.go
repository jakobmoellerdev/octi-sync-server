package redis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"

	"github.com/go-redis/redis/v9"
)

func NewModules(client *redis.Client) *Modules {
	return &Modules{client}
}

type Modules struct {
	client *redis.Client
}

func (r *Modules) Set(ctx context.Context, name string, module service.Module) error {
	moduleData, err := io.ReadAll(module.Raw())
	if err == nil {
		NoExpiry := time.Duration(-1)
		err = r.client.Set(ctx, name, moduleData, NoExpiry).Err()
	}
	if err != nil {
		return fmt.Errorf("persisting %s failed: %w", name, service.ErrWritingModuleFailed)
	}
	return nil
}

func (r *Modules) Get(ctx context.Context, name string) (service.Module, error) {
	bytes, err := r.client.Get(ctx, name).Bytes()
	if errors.Is(err, redis.Nil) {
		return ModuleFromBytes([]byte{}), nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading %s failed: %w", name, service.ErrReadingModule)
	}

	return ModuleFromBytes(bytes), nil
}

func (r *Modules) HealthCheck() service.HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "redis-modules", r.client.Ping(ctx).Err() == nil
	}
}
