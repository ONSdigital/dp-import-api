package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	datasetclient "github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/mongo"
	kafka "github.com/ONSdigital/dp-kafka"

	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/dataset"
	"github.com/ONSdigital/dp-import-api/importqueue"
	"github.com/ONSdigital/dp-import-api/initialise"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/recipe"
	"github.com/ONSdigital/dp-import-api/url"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
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

	dataBakerProducer, err := serviceList.GetProducer(ctx, cfg.Brokers, cfg.DatabakerImportTopic, initialise.DataBaker, cfg.KafkaMaxBytes)
	logIfError(ctx, err, "databaker kafka producer error")
	dataBakerProducer.Channels().LogErrors(ctx, "error received from kafka data baker producer, topic: "+cfg.DatabakerImportTopic)

	inputFileAvailableProducer, err := serviceList.GetProducer(ctx, cfg.Brokers, cfg.InputFileAvailableTopic, initialise.Direct, cfg.KafkaMaxBytes)
	logIfError(ctx, err, "direct kafka producer error")
	inputFileAvailableProducer.Channels().LogErrors(ctx, "error received from kafka input file available producer, topic: "+cfg.InputFileAvailableTopic)

	urlBuilder := url.NewBuilder(cfg.Host, cfg.DatasetAPIURL)
	jobQueue := importqueue.CreateImportQueue(dataBakerProducer.Channels().Output, inputFileAvailableProducer.Channels().Output)

	client := dphttp.NewClient()
	datasetAPI := dataset.API{Client: client, URL: cfg.DatasetAPIURL, ServiceAuthToken: cfg.ServiceAuthToken}
	recipeAPI := recipe.API{Client: client, URL: cfg.RecipeAPIURL}

	jobService := job.NewService(mongoDataStore, jobQueue, &datasetAPI, &recipeAPI, urlBuilder)

	hc := startHealthChecks(ctx, cfg, dataBakerProducer, inputFileAvailableProducer, mongoDataStore)

	api.CreateImportAPI(ctx, cfg.BindAddr, cfg.ZebedeeURL, mongoDataStore, jobService, hc)

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

		hc.Stop()

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
			logIfError(ctx, inputFileAvailableProducer.Close(ctx), "unable to close direct producer")
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	log.Event(ctx, "Shutdown complete", log.INFO)
	os.Exit(1)
}

func startHealthChecks(ctx context.Context,
	cfg *config.Configuration,
	dataBakerProducer *kafka.Producer,
	inputFileAvailableProducer *kafka.Producer,
	mongoDataStore *mongo.Mongo) *healthcheck.HealthCheck {

	hasErrors := false
	versionInfo, err := healthcheck.NewVersionInfo(BuildTime, GitCommit, Version)
	if err != nil {
		log.Event(ctx, "error creating version info", log.FATAL, log.Error(err))
		hasErrors = true
	}

	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	if err = hc.AddCheck("Kafka Data Baker Producer", dataBakerProducer.Checker); err != nil {
		log.Event(ctx, "error adding check for kafka data baker producer", log.ERROR, log.Error(err))
		hasErrors = true
	}

	if err = hc.AddCheck("Kafka Input File Available Producer", inputFileAvailableProducer.Checker); err != nil {
		log.Event(ctx, "error adding check for kafka input file available producer", log.ERROR, log.Error(err))
		hasErrors = true
	}

	zebedeeClient := zebedee.New(cfg.ZebedeeURL)
	if err = hc.AddCheck("Zebedee", zebedeeClient.Checker); err != nil {
		log.Event(ctx, "error creating zebedee health check", log.ERROR, log.Error(err))
		hasErrors = true
	}

	checkMongoClient := mongoDataStore.HealthCheckClient()
	if err = hc.AddCheck("MongoDB", checkMongoClient.Checker); err != nil {
		log.Event(ctx, "error creating mongodb health check", log.ERROR, log.Error(err))
		hasErrors = true
	}

	datasetAPIClient := datasetclient.NewAPIClient(cfg.DatasetAPIURL)
	if err = hc.AddCheck("Dataset API", datasetAPIClient.Checker); err != nil {
		log.Event(ctx, "error creating dataset API health check", log.Error(err))
		hasErrors = true
	}

	recipeAPIHealthCheckClient := health.NewClient("Recipe API", cfg.RecipeAPIURL)
	if err = hc.AddCheck("Recipe API", recipeAPIHealthCheckClient.Checker); err != nil {
		log.Event(ctx, "error creating recipe API health check", log.Error(err))
		hasErrors = true
	}

	if hasErrors {
		os.Exit(1)
	}

	hc.Start(ctx)

	return &hc
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
