package redis_test

import (
	"testing"

	"github.com/jakob-moeller-cloud/octi-sync-server/service/redis"
	"github.com/stretchr/testify/assert"
)

func Test_AccountFromUsername(t *testing.T) {
	t.Parallel()
	assertions := assert.New(t)
	acc := redis.AccountFromUserAndHashedPass("test", "pass")
	assertions.Equal("test", acc.Username())
	assertions.Equal("pass", acc.HashedPass())
}
