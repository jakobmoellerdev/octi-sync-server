package main

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/server"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/rs/zerolog"
	baseLogger "github.com/rs/zerolog/log"
	"github.com/sethvargo/go-password/password"
)

var (
	version = "dev"
	commit  = "none"    //nolint:gochecknoglobals
	date    = "unknown" //nolint:gochecknoglobals
	builtBy = "unknown" //nolint:gochecknoglobals
)

// Func main should be as small as possible and do as little as possible by convention.
func main() {
	// Generate our config based on the config supplied
	// by the user in the flags
	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var logger zerolog.Logger

	switch cfg.LogSettings.Format {
	case config.LogSettingsFormatPretty:
		logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	case config.LogSettingsFormatJSON:
		fallthrough
	case config.LogSettingsFormatNone:
		fallthrough
	default:
		logger = baseLogger.With().Logger()
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		logger = logger.With().Str("build", info.Main.Version).Logger()
		logger.Info().Msg(fmt.Sprintf("Welcome to Octi: %s, commit %s, built at %s by %s", version, commit, date, builtBy))
	}

	cfg.Logger = &logger

	cfg.PasswordGenerator, err = password.NewGenerator(nil)

	if err != nil {
		log.Fatal(err)
	}

	cfg.UsernameGenerator, err = service.NewUsernameGenerator(service.UUIDUsernameGeneration)

	if err != nil {
		log.Fatal(err)
	}

	uuid.EnableRandPool()

	// Run the server
	if err := server.Run(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
}
