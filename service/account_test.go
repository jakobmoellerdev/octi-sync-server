package service_test

import (
	"testing"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/stretchr/testify/assert"
)

func Test_AccountFromUsername(t *testing.T) {
	t.Parallel()
	assertions := assert.New(t)
	acc := service.NewBaseAccount("test", "pass")
	assertions.Equal("test", acc.Username())
	assertions.Equal("pass", acc.HashedPass())
}
