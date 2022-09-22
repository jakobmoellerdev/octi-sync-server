package config_test

import (
	"flag"
	"sync"
	"testing"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/stretchr/testify/assert"
)

func Test_ReadConfigFromFile(t *testing.T) {
	t.Parallel()

	c, err := config.NewConfig("./test_config.yaml")
	assertions := assert.New(t)

	assertions.NoError(err)
	assertions.Equal("127.0.0.1", c.Server.Host)
	assertions.Equal("8080", c.Server.Port)
}

func Test_ReadConfigFromFile_NotFound(t *testing.T) {
	t.Parallel()

	_, err := config.NewConfig("./xxx.yaml")
	assertions := assert.New(t)

	assertions.Error(err)
	assertions.ErrorContains(err, "config cannot be opened")
}

func Test_ReadConfigFromFile_Invalid(t *testing.T) {
	t.Parallel()

	_, err := config.NewConfig("./invalid_config.yaml")
	assertions := assert.New(t)

	assertions.Error(err)
	assertions.ErrorContains(err, "config cannot be decoded")
}

func Test_ValidateConfigPath(t *testing.T) {
	t.Parallel()

	err := config.ValidateConfigPath("./test_config.yaml")
	assertions := assert.New(t)

	assertions.NoError(err)
}

func Test_ValidateConfigPath_Invalid(t *testing.T) {
	t.Parallel()

	err := config.ValidateConfigPath("./xxx.yaml")
	assertions := assert.New(t)

	assertions.Error(err)
	assertions.ErrorContains(err, "config path invalid")
}

func Test_ValidateConfigPath_Invalid_Dir(t *testing.T) {
	t.Parallel()

	err := config.ValidateConfigPath(t.TempDir())

	assertions := assert.New(t)

	assertions.Error(err)
	assertions.ErrorIs(err, config.ErrIsADirectory)
}

// necessary to avoid data race in flag tests.
var flagSync = sync.Mutex{} //nolint:gochecknoglobals

func Test_ParseFlags(t *testing.T) {
	t.Parallel()

	flagSync.Lock()
	defer flagSync.Unlock()
	flag.Set("config", "test_config.yaml")

	_, err := config.ParseFlags()

	assertions := assert.New(t)

	assertions.NoError(err)
}

func Test_ParseFlags_Invalid(t *testing.T) {
	t.Parallel()

	flagSync.Lock()
	defer flagSync.Unlock()
	flag.Set("config", "xxx")

	_, err := config.ParseFlags()

	assertions := assert.New(t)

	assertions.Error(err)
}
