package main

import (
	"context"
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
	"github.com/ONSdigital/log.go/log"
)

const serviceNamespace = "dp-import-api"

func main() {

	ctx := context.Background()
	log.Namespace = serviceNamespace

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	cfg, err := config.Get()
	exitIfError(ctx, err, "unable to retrieve configuration")

	// sensitive fields are omitted from config.String().
	log.Event(ctx, "loaded config", log.INFO, log.Data{
		"config": cfg,
	})

	var serviceList initialise.ExternalServiceList

	mongoDataStore, err := serviceList.GetMongoDataStore(cfg)
	logIfError(ctx, err, "mongodb datastore error")

	dataBakerProducer, err := serviceList.GetProducer(cfg.Brokers, cfg.DatabakerImportTopic, initialise.DataBaker, cfg.KafkaMaxBytes)
	logIfError(ctx, err, "databaker kafka producer error")

	directProducer, err := serviceList.GetProducer(cfg.Brokers, cfg.InputFileAvailableTopic, initialise.Direct, cfg.KafkaMaxBytes)
	logIfError(ctx, err, "direct kafka producer error")

	auditProducer, err := serviceList.GetProducer(cfg.Brokers, cfg.AuditEventsTopic, initialise.Audit, cfg.KafkaMaxBytes)
	logIfError(ctx, err, "direct kafka producer error")

	urlBuilder := url.NewBuilder(cfg.Host, cfg.DatasetAPIURL)
	jobQueue := importqueue.CreateImportQueue(dataBakerProducer.Output(), directProducer.Output())

	client := rchttp.NewClient()
	datasetAPI := dataset.API{Client: client, URL: cfg.DatasetAPIURL, ServiceAuthToken: cfg.ServiceAuthToken}
	recipeAPI := recipe.API{Client: client, URL: cfg.RecipeAPIURL}

	jobService := job.NewService(mongoDataStore, jobQueue, &datasetAPI, &recipeAPI, urlBuilder)
	auditor := audit.New(auditProducer, serviceNamespace)

	api.CreateImportAPI(ctx, cfg.BindAddr, cfg.ZebedeeURL, mongoDataStore, jobService, auditor)

	// TODO - Discussion on handling kafka channels and how we would implement reconnecting to kafka
	// Function below is to handle errors received by kafka as gracefully as possible without panicing or
	// continuously logging the same error messages - will likely be replaced when kafka reconnects are implemented
	go func() {
		var databaKerProducerErrors, directProducerErrors, auditProducerErrors chan (error)

		if serviceList.DataBakerProducer {
			databaKerProducerErrors = dataBakerProducer.Errors()
		} else {
			databaKerProducerErrors = make(chan error, 1)
		}

		if serviceList.DirectProducer {
			directProducerErrors = directProducer.Errors()
		} else {
			directProducerErrors = make(chan error, 1)
		}

		if serviceList.AuditProducer {
			auditProducerErrors = auditProducer.Errors()
		} else {
			auditProducerErrors = make(chan error, 1)
		}

		select {
		case err := <-databaKerProducerErrors:
			log.Event(ctx, "kafka databaker producer", log.ERROR, log.Error(err))
		case err := <-directProducerErrors:
			log.Event(ctx, "kafka direct producer", log.ERROR, log.Error(err))
		case err := <-auditProducerErrors:
			log.Event(ctx, "kafka audit producer", log.ERROR, log.Error(err))
		}
	}()

	// block until a fatal error occurs
	select {
	case sig := <-signals:
		log.Event(ctx, "os signal received", log.INFO, log.Data{"signal": sig.String()})
	}

	log.Event(ctx, "Shutdown service", log.INFO, log.Data{"timeout": cfg.GracefulShutdownTimeout})
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		log.Event(ctx, "closing http server for healthcheck", log.INFO)
		if err = api.Close(ctx); err != nil {
			logIfError(ctx, err, "unable to close api server")
		}

		if serviceList.MongoDataStore {
			log.Event(ctx, "closing mongo data store", log.INFO)
			// mongo.Close() may use all remaining time in the context
			logIfError(ctx, mongoDataStore.Close(ctx), "unable to close mongo data store")
		}

		if serviceList.DataBakerProducer {
			log.Event(ctx, "closing data baker producer", log.INFO)
			logIfError(ctx, dataBakerProducer.Close(ctx), "unable to close data baker producer")
		}

		if serviceList.DirectProducer {
			log.Event(ctx, "closing direct producer", log.INFO)
			logIfError(ctx, directProducer.Close(ctx), "unable to close direct producer")
		}

		// Close audit producer last, to maximise capturing all events
		if serviceList.AuditProducer {
			log.Event(ctx, "closing audit producer", log.INFO)
			logIfError(ctx, auditProducer.Close(ctx), "unable to close audit producer")
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	log.Event(ctx, "Shutdown complete", log.INFO)
	os.Exit(1)
}

func exitIfError(ctx context.Context, err error, message string) {
	if err != nil {
		log.Event(ctx, message, log.Error(err))
		os.Exit(1)
	}
}

func logIfError(ctx context.Context, err error, message string) {
	if err != nil {
		log.Event(ctx, message, log.Error(err))
	}
}
