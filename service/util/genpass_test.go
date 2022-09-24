package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jakob-moeller-cloud/octi-sync-server/service/util"
)

func TestGenPass(t *testing.T) {
	t.Parallel()

	pw := util.NewInPlacePasswordGenerator().
		Generate(32, 6, 6, 6)

	assert.New(t).Equal(32, len(pw))
}
