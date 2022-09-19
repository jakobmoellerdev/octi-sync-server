package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-redis/redis/v9"
)

func NewRedisModules(client *redis.Client) *RedisModules {
	return &RedisModules{client}
}

type RedisModules struct {
	client *redis.Client
}

func (r *RedisModules) Set(ctx context.Context, name string, module Module) error {
	moduleData, err := io.ReadAll(module.Raw())
	if err == nil {
		NoExpiry := time.Duration(-1)
		err = r.client.Set(ctx, name, moduleData, NoExpiry).Err()
	}
	if err != nil {
		return fmt.Errorf("persisting %s failed: %w", name, ErrWritingModuleFailed)
	}
	return nil
}

func (r *RedisModules) Get(ctx context.Context, name string) (Module, error) {
	bytes, err := r.client.Get(ctx, name).Bytes()
	if errors.Is(err, redis.Nil) {
		return RedisModuleFromBytes([]byte{}), nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading %s failed: %w", name, ErrReadingModule)
	}

	return RedisModuleFromBytes(bytes), nil
}
