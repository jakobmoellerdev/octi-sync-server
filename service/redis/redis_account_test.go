package redis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jakob-moeller-cloud/octi-sync-server/service/redis"
)

func Test_AccountFromUsername(t *testing.T) {
	t.Parallel()
	assertions := assert.New(t)
	acc := redis.AccountFromUserAndHashedPass("test", "pass")
	assertions.Equal("test", acc.Username())
	assertions.Equal("pass", acc.HashedPass())
}
