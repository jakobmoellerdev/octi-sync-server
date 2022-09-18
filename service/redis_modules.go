package service

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
	"io"
	"time"
)

var noExpiry = time.Duration(-1)

func NewRedisModules(client *redis.Client) *RedisModules {
	return &RedisModules{client}
}

type RedisModules struct {
	client *redis.Client
}

func (r *RedisModules) Set(ctx context.Context, name string, module Module) error {
	moduleData, err := io.ReadAll(module.Raw())
	if err == nil {
		err = r.client.Set(ctx, name, moduleData, noExpiry).Err()
	}
	if err != nil {
		return fmt.Errorf("persisting %s failed: %w", name, ErrWritingModuleFailed)
	}
	return nil
}

func (r *RedisModules) Get(ctx context.Context, name string) (Module, error) {
	bytes, err := r.client.Get(ctx, name).Bytes()
	if err == redis.Nil {
		return RedisModuleFromBytes([]byte{}), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s failed: %w", name, ErrReadingModule)
	}
	return RedisModuleFromBytes(bytes), nil
}
