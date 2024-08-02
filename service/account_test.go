package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jakobmoellerdev/octi-sync-server/service"
)

func Test_AccountFromUsername(t *testing.T) {
	t.Parallel()
	assertions := assert.New(t)
	acc := service.NewBaseAccount("test", time.Now())
	assertions.Equal("test", acc.Username())
}
