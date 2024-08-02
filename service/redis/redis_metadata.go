package redis

import (
	"context"
	"errors"
	"fmt"

	json "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"

	"github.com/jakobmoellerdev/octi-sync-server/service"
)

const MetadataKeySpace = "octi:metadata"

type MetadataProvider struct {
	Client redis.Cmdable
}

func (r *MetadataProvider) metadataKey(id service.MetadataID) string {
	return fmt.Sprintf("%s:%s", MetadataKeySpace, id)
}

func (r *MetadataProvider) Get(ctx context.Context, id service.MetadataID) (service.Metadata, error) {
	name := r.metadataKey(id)

	bytes, err := r.Client.Get(ctx, name).Bytes()

	if errors.Is(err, redis.Nil) {
		return nil, service.ErrNoMetadata
	}

	var metaData service.BaseMetadata
	if err := json.Unmarshal(bytes, &metaData); err != nil {
		return nil, fmt.Errorf("unmarshalling meta %s failed: %w", name, service.ErrWritingModuleFailed)
	}

	return &metaData, nil
}

func (r *MetadataProvider) Set(ctx context.Context, meta service.Metadata) error {
	name := r.metadataKey(meta.GetID())

	data, err := json.Marshal(&meta)
	if err != nil {
		return fmt.Errorf("marshalling meta %s failed: %w", name, service.ErrWritingModuleFailed)
	}

	err = r.Client.Set(ctx, name, data, NoExpiry).Err()
	if err != nil {
		return fmt.Errorf("persisting meta %s failed: %w", name, service.ErrWritingModuleFailed)
	}

	return nil
}

func (r *MetadataProvider) HealthCheck() service.HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "redis-metadata-provider", r.Client.Ping(ctx).Err() == nil
	}
}
