package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/dataset"
	"github.com/ONSdigital/dp-import-api/importqueue"
	"github.com/ONSdigital/dp-import-api/initialise"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/recipe"
	"github.com/ONSdigital/dp-import-api/url"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/log"
)

const serviceNamespace = "dp-import-api"

func main() {
	log.Namespace = serviceNamespace

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	cfg, err := config.Get()
	exitIfError(err, "unable to retrieve configuration")

	// sensitive fields are omitted from config.String().
	log.Info("loaded config", log.Data{
		"config": cfg,
	})

	var serviceList initialise.ExternalServiceList

	mongoDataStore, err := serviceList.GetMongoDataStore(cfg)
	logIfError(err, "mongodb datastore error")

	dataBakerProducer, err := serviceList.GetProducer(cfg.Brokers, cfg.DatabakerImportTopic, initialise.DataBakerProducer, cfg.KafkaMaxBytes)
	logIfError(err, "databaker kafka producer error")

	directProducer, err := serviceList.GetProducer(cfg.Brokers, cfg.InputFileAvailableTopic, initialise.DirectProducer, cfg.KafkaMaxBytes)
	logIfError(err, "direct kafka producer error")

	auditProducer, err := serviceList.GetProducer(cfg.Brokers, cfg.AuditEventsTopic, initialise.AuditProducer, cfg.KafkaMaxBytes)
	logIfError(err, "direct kafka producer error")

	urlBuilder := url.NewBuilder(cfg.Host, cfg.DatasetAPIURL)
	jobQueue := importqueue.CreateImportQueue(dataBakerProducer.Output(), directProducer.Output())

	client := rchttp.NewClient()
	datasetAPI := dataset.API{Client: client, URL: cfg.DatasetAPIURL, ServiceAuthToken: cfg.ServiceAuthToken}
	recipeAPI := recipe.API{Client: client, URL: cfg.RecipeAPIURL}

	jobService := job.NewService(mongoDataStore, jobQueue, &datasetAPI, &recipeAPI, urlBuilder)
	auditor := audit.New(auditProducer, serviceNamespace)

	api.CreateImportAPI(cfg.Host, cfg.BindAddr, cfg.ZebedeeURL, mongoDataStore, jobService, auditor)

	go func() {
		select {
		case err := <-dataBakerProducer.Errors():
			log.ErrorC("kafka databaker producer", err, nil)
		case err := <-directProducer.Errors():
			log.ErrorC("kafka direct producer", err, nil)
		case err := <-auditProducer.Errors():
			log.ErrorC("kafka audit producer", err, nil)
		}
	}()

	// block until a fatal error occurs
	select {
	case sig := <-signals:
		log.Error(errors.New("os signal received"), log.Data{"signal": sig.String()})
	}

	log.Info("Shutdown service", log.Data{"timeout": cfg.GracefulShutdownTimeout})
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		log.Info("closing http server for healthcheck", nil)
		if err = api.Close(ctx); err != nil {
			logIfError(err, "unable to close api server")
		}

		if serviceList.MongoDataStore {
			log.Info("closing mongo data store", nil)
			// mongo.Close() may use all remaining time in the context
			logIfError(mongoDataStore.Close(ctx), "unable to close mongo data store")
		}

		if serviceList.DataBakerProducer {
			log.Info("closing data baker producer", nil)
			logIfError(dataBakerProducer.Close(ctx), "unable to close data baker producer")
		}

		if serviceList.DirectProducer {
			log.Info("closing direct producer", nil)
			logIfError(directProducer.Close(ctx), "unable to close direct producer")
		}

		// Close audit producer last, to maximise capturing all events
		if serviceList.AuditProducer {
			log.Info("closing audit producer", nil)
			logIfError(auditProducer.Close(ctx), "unable to close audit producer")
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	log.Info("Shutdown complete", nil)
	os.Exit(1)
}

func exitIfError(err error, message string) {
	if err != nil {
		log.ErrorC(message, err, nil)
		os.Exit(1)
	}
}

func logIfError(err error, message string) {
	if err != nil {
		log.ErrorC(message, err, nil)
	}
}
