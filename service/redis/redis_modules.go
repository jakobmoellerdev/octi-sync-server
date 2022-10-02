package redis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/util"
)

type Modules struct {
	Client     redis.Cmdable
	Expiration time.Duration
}

func (r *Modules) Set(ctx context.Context, name string, module service.Module) error {
	moduleData, err := io.ReadAll(module.Raw())
	if err == nil {
		err = r.Client.Set(ctx, name, moduleData, r.Expiration).Err()
	}

	if err != nil {
		return fmt.Errorf("persisting %s failed: %w", name, service.ErrWritingModuleFailed)
	}

	return nil
}

func (r *Modules) Get(ctx context.Context, name string) (service.Module, error) {
	bytes, err := r.Client.Get(ctx, name).Bytes()
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
		return "redis-modules", r.Client.Ping(ctx).Err() == nil
	}
}

func (r *Modules) DeleteByPattern(ctx context.Context, pattern string) error {
	keys, err := r.Client.Keys(ctx, pattern).Result()

	if err != nil {
		return fmt.Errorf("error while getting keys with pattern %s, %w", pattern, err)
	}

	var errs []error

	for i := range keys {
		if err := r.Delete(ctx, keys[i]); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return util.MultiError(errs)
	}

	return nil
}

func (r *Modules) Delete(ctx context.Context, key string) error {
	err := r.Client.Del(ctx, key).Err()

	if err != nil {
		return fmt.Errorf("error while deleting %s: %w", key, err)
	}

	return nil
}
