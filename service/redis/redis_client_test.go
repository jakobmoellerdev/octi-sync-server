package redis_test

import (
	"context"
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/jakobmoellerdev/octi-sync-server/config"
	"github.com/jakobmoellerdev/octi-sync-server/service/redis"
	"github.com/jakobmoellerdev/octi-sync-server/service/redis/mock"
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
