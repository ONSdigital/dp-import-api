package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/service"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

const serviceName = "dp-import-api"

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Error(ctx, "application unexpectedly failed", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	svcErrors := make(chan error, 1)

	// Read config
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "unable to retrieve configuration", err)
		return err
	}

	// sensitive fields are omitted from config.String().
	log.Info(ctx, "loaded config", log.Data{
		"config": cfg,
	})

	// Run the service
	svc := service.New()
	if err := svc.Init(ctx, cfg, BuildTime, GitCommit, Version); err != nil {
		return errors.Wrap(err, "running service failed")
	}
	svc.Start(ctx, svcErrors)

	// Blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-svcErrors:
		log.Error(ctx, "service error received", err)
	case sig := <-signals:
		log.Info(ctx, "os signal received", log.Data{"signal": sig})
	}
	return svc.Close(ctx)
}
