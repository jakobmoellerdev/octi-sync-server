package redis_test

import (
	"context"
	"testing"

	goredis "github.com/go-redis/redis/v9"
	"github.com/golang/mock/gomock"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/redis"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/redis/mock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func Test_NewClientWithRegularPing(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	clientMock := mock.NewMockUniversalClient(ctrl)
	log := zerolog.New(zerolog.NewTestWriter(t))
	cfg := &config.Config{Logger: &log}

	clients, err := redis.NewClientsWithRegularPing(
		context.Background(), cfg,
		func(config *config.Config) goredis.UniversalClient {
			return clientMock
		}, redis.ClientMutators{},
	)
	assert.New(t).NoError(err)
	assert.NotEmpty(t, clients)
	assert.Len(t, clients, 1)
}
