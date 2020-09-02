package service

import (
	"context"

	clientshealth "github.com/ONSdigital/dp-api-clients-go/health"
	clientsidentity "github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/dp-api-clients-go/middleware"
	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/dataset"
	"github.com/ONSdigital/dp-import-api/importqueue"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/mongo"
	"github.com/ONSdigital/dp-import-api/recipe"
	"github.com/ONSdigital/dp-import-api/url"
	kafka "github.com/ONSdigital/dp-kafka"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the Dataset API
type Service struct {
	cfg                        *config.Configuration
	serviceList                *ExternalServiceList
	mongoDataStore             *mongo.Mongo
	dataBakerProducer          kafka.IProducer
	inputFileAvailableProducer kafka.IProducer
	server                     HTTPServer
	healthCheck                HealthChecker
	importAPI                  *api.ImportAPI
	identityClient             *clientsidentity.Client
	datasetAPI                 *dataset.API
	recipeAPI                  *recipe.API
}

// New creates a new service
func New(cfg *config.Configuration, serviceList *ExternalServiceList) *Service {
	svc := &Service{
		cfg:         cfg,
		serviceList: serviceList,
	}
	return svc
}

// Run the service
func (svc *Service) Run(ctx context.Context, buildTime, gitCommit, version string, svcErrors chan error) (err error) {

	// Get mongoDB connection
	svc.mongoDataStore, err = svc.serviceList.GetMongoDataStore(svc.cfg)
	logIfError(ctx, err, "mongodb datastore error")

	// Get data baker kafka producer
	svc.dataBakerProducer, err = svc.serviceList.GetProducer(ctx, svc.cfg.Brokers, svc.cfg.DatabakerImportTopic, DataBaker, svc.cfg.KafkaMaxBytes)
	logIfError(ctx, err, "databaker kafka producer error")

	// Get input file available kafka producer
	svc.inputFileAvailableProducer, err = svc.serviceList.GetProducer(ctx, svc.cfg.Brokers, svc.cfg.InputFileAvailableTopic, Direct, svc.cfg.KafkaMaxBytes)
	logIfError(ctx, err, "direct kafka producer error")

	// Create Identity Client
	svc.identityClient = clientsidentity.New(svc.cfg.ZebedeeURL)

	// Create dataset and recie API clients.
	// TODO: We should consider replacing these with the corresponding dp-api-clients-go clients
	client := dphttp.NewClient()
	svc.datasetAPI = &dataset.API{Client: client, URL: svc.cfg.DatasetAPIURL, ServiceAuthToken: svc.cfg.ServiceAuthToken}
	svc.recipeAPI = &recipe.API{Client: client, URL: svc.cfg.RecipeAPIURL}

	// Get HealthCheck
	svc.healthCheck, err = svc.serviceList.GetHealthCheck(svc.cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return err
	}
	if err := svc.registerCheckers(ctx); err != nil {
		return errors.Wrap(err, "unable to register checkers")
	}

	// Get HTTP router and server with middleware
	r := mux.NewRouter()
	m := svc.createMiddleware(svc.cfg) // TODO implement
	svc.server = svc.serviceList.GetHTTPServer(svc.cfg.BindAddr, m.Then(r))

	// Create API with job service
	urlBuilder := url.NewBuilder(svc.cfg.Host, svc.cfg.DatasetAPIURL)
	jobQueue := importqueue.CreateImportQueue(svc.dataBakerProducer.Channels().Output, svc.inputFileAvailableProducer.Channels().Output)
	jobService := job.NewService(svc.mongoDataStore, jobQueue, svc.datasetAPI, svc.recipeAPI, urlBuilder)
	svc.importAPI = api.Setup(r, svc.mongoDataStore, jobService)

	// Start kafka logging
	svc.dataBakerProducer.Channels().LogErrors(ctx, "error received from kafka data baker producer, topic: "+svc.cfg.DatabakerImportTopic)
	svc.inputFileAvailableProducer.Channels().LogErrors(ctx, "error received from kafka input file available producer, topic: "+svc.cfg.InputFileAvailableTopic)

	svc.healthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		log.Event(ctx, "Starting api...", log.INFO)
		if err := svc.server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return nil
}

// CreateMiddleware creates an Alice middleware chain of handlers
// to forward collectionID from cookie from header
func (svc *Service) createMiddleware(cfg *config.Configuration) alice.Chain {
	identityHandler := dphandlers.IdentityWithHTTPClient(svc.identityClient)
	return alice.New(
		middleware.Whitelist(middleware.HealthcheckFilter(svc.healthCheck.Handler)),
		dprequest.HandlerRequestID(16),
		identityHandler,
	)
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.cfg.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.serviceList.HealthCheck {
			svc.healthCheck.Stop()
		}

		// stop any incoming requests
		if logIfError(ctx, svc.server.Shutdown(ctx), "failed to shutdown http server") {
			hasShutdownError = true
		}

		// close API
		if logIfError(ctx, svc.importAPI.Close(ctx), "error closing API") {
			hasShutdownError = true
		}

		// Close MongoDB (if it exists)
		if svc.serviceList.MongoDataStore {
			log.Event(ctx, "closing mongo data store", log.INFO)
			if logIfError(ctx, svc.mongoDataStore.Close(ctx), "unable to close mongo data store") {
				hasShutdownError = true
			}
		}

		// Close Data Baker Kafka Producer (it if exists)
		if svc.serviceList.DataBakerProducer {
			log.Event(ctx, "closing data baker producer", log.INFO)
			if logIfError(ctx, svc.dataBakerProducer.Close(ctx), "unable to close data baker producer") {
				hasShutdownError = true
			}
		}

		// Close Direct Kafka Producer (if it exists)
		if svc.serviceList.DirectProducer {
			log.Event(ctx, "closing direct producer", log.INFO)
			if logIfError(ctx, svc.inputFileAvailableProducer.Close(ctx), "unable to close direct producer") {
				hasShutdownError = true
			}
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "shutdown timed out", log.ERROR, log.Error(ctx.Err()))
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

// registerCheckers adds the checkers for the service clients to the health check object
func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false

	if err = svc.healthCheck.AddCheck("Kafka Data Baker Producer", svc.dataBakerProducer.Checker); err != nil {
		log.Event(ctx, "error adding check for kafka data baker producer", log.ERROR, log.Error(err))
		hasErrors = true
	}

	if err = svc.healthCheck.AddCheck("Kafka Input File Available Producer", svc.inputFileAvailableProducer.Checker); err != nil {
		log.Event(ctx, "error adding check for kafka input file available producer", log.ERROR, log.Error(err))
		hasErrors = true
	}

	if err = svc.healthCheck.AddCheck("Zebedee", svc.identityClient.Checker); err != nil {
		log.Event(ctx, "error adding checker for zebedee", log.ERROR, log.Error(err))
		hasErrors = true
	}

	checkMongoClient := svc.mongoDataStore.HealthCheckClient()
	if err = svc.healthCheck.AddCheck("MongoDB", checkMongoClient.Checker); err != nil {
		log.Event(ctx, "error creating mongodb health check", log.ERROR, log.Error(err))
		hasErrors = true
	}

	datasetAPIClient := clientshealth.NewClientWithClienter("dataset-api", svc.datasetAPI.URL, svc.datasetAPI.Client)
	if err = svc.healthCheck.AddCheck("Dataset API", datasetAPIClient.Checker); err != nil {
		log.Event(ctx, "error creating dataset API health check", log.Error(err))
		hasErrors = true
	}

	recipeAPIHealthCheckClient := clientshealth.NewClientWithClienter("recipe-api", svc.recipeAPI.URL, svc.recipeAPI.Client)
	if err = svc.healthCheck.AddCheck("Recipe API", recipeAPIHealthCheckClient.Checker); err != nil {
		log.Event(ctx, "error creating recipe API health check", log.Error(err))
		hasErrors = true
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}

func logIfError(ctx context.Context, err error, message string) bool {
	if err != nil {
		log.Event(ctx, message, log.ERROR, log.Error(err))
		return true
	}
	return false
}
