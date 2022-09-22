package v1_test

import (
	"context"
	"testing"

	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	log := zerolog.New(zerolog.NewTestWriter(t))
	v1.New(context.Background(), echo.New(), &config.Config{
		Logger: &log,
	})
}
