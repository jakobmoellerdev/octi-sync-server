package service_test

import (
	"testing"
	"time"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/stretchr/testify/assert"
)

func Test_AccountFromUsername(t *testing.T) {
	t.Parallel()
	assertions := assert.New(t)
	acc := service.NewBaseAccount("test", time.Now())
	assertions.Equal("test", acc.Username())
}
