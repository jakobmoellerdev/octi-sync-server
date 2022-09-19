package main

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"log"
	"net/http"
	"octi-sync-server/config"
	"octi-sync-server/redis"
	"octi-sync-server/router"
	"octi-sync-server/service"
	"os"
	"os/signal"
	"time"
)

// Func main should be as small as possible and do as little as possible by convention
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

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}(logger)

	cfg.Logger = logger

	uuid.EnableRandPool()

	// Run the server
	Run(cfg)
}

// Run will run the HTTP Server
func Run(config *config.Config) {
	startUpContext, cancelStartUpContext := context.WithCancel(context.Background())
	defer cancelStartUpContext()

	client, err := redis.NewClientWithRegularPing(startUpContext, config)
	if err != nil {
		log.Fatal(err)
	}
	config.Services.Accounts = service.Accounts(service.NewRedisAccounts(client))
	config.Services.Modules = service.Modules(service.NewRedisModules(client))
	config.Services.Devices = service.Devices(service.NewRedisDevices(client))

	// Define server options
	srv := &http.Server{
		Addr:         config.Server.Host + ":" + config.Server.Port,
		Handler:      router.New(startUpContext, config),
		ReadTimeout:  config.Server.Timeout.Read * time.Second,
		WriteTimeout: config.Server.Timeout.Write * time.Second,
		IdleTimeout:  config.Server.Timeout.Idle * time.Second,
	}

	idleConsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint // We received an interrupt signal, shut down.
		// Set up a context to allow for graceful server shutdowns in the event
		// of an OS interrupt (defers the cancel just in case)
		ctx, cancel := context.WithTimeout(
			startUpContext,
			config.Server.Timeout.Server,
		)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			config.Logger.Warn("server shutdown error: " + err.Error())
		}
		close(idleConsClosed)
	}()

	// service connections
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		config.Logger.Fatal("listen: %s" + err.Error())
	}

	<-idleConsClosed
	config.Logger.Info("server shut down finished")
}
