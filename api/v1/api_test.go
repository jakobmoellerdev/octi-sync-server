package v1_test

import (
	"context"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	v1 "github.com/jakobmoellerdev/octi-sync-server/api/v1"
	"github.com/jakobmoellerdev/octi-sync-server/config"
)

func TestNew(t *testing.T) {
	log := zerolog.New(zerolog.NewTestWriter(t))
	v1.New(context.Background(), echo.New(), &config.Config{
		Logger: &log,
	})
}
