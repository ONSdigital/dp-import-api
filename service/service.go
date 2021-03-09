package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	clientshealth "github.com/ONSdigital/dp-api-clients-go/health"
	clientsidentity "github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/dp-api-clients-go/middleware"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/dataset"
	"github.com/ONSdigital/dp-import-api/datastore"
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
	mongoDataStore             datastore.DataStorer
	dataBakerProducer          kafka.IProducer
	inputFileAvailableProducer kafka.IProducer
	server                     HTTPServer
	healthCheck                HealthChecker
	importAPI                  *api.ImportAPI
	identityClient             *clientsidentity.Client
	datasetAPI                 *dataset.API
	recipeAPI                  *recipe.API
}

// getMongoDataStore creates a mongoDB connection
var getMongoDataStore = func(cfg *config.Configuration) (datastore.DataStorer, error) {
	return mongo.NewDatastore(cfg.MongoDBURL, cfg.MongoDBDatabase, cfg.MongoDBCollection)
}

// getKafkaProducer creates a new Kafka Producer
var getKafkaProducer = func(ctx context.Context, kafkaBrokers []string, topic string, envMax int) (kafka.IProducer, error) {
	producerChannels := kafka.CreateProducerChannels()
	return kafka.NewProducer(ctx, kafkaBrokers, topic, envMax, producerChannels)
}

// getHealthCheck returns a healthcheck
var getHealthCheck = func(version healthcheck.VersionInfo, criticalTimeout, interval time.Duration) HealthChecker {
	hc := healthcheck.New(version, criticalTimeout, interval)
	return &hc
}

// getHTTPServer returns an http server
var getHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// New creates a new service
func New() *Service {
	return &Service{}
}

// Init initialises all the service dependencies, including healthcheck with checkers, api and middleware.
func (svc *Service) Init(ctx context.Context, cfg *config.Configuration, buildTime, gitCommit, version string) (err error) {

	svc.cfg = cfg

	// Get mongoDB connection (non-fatal)
	svc.mongoDataStore, err = getMongoDataStore(svc.cfg)
	if err != nil {
		log.Event(ctx, "mongodb datastore error", log.ERROR, log.Error(err))
	}

	// Get data baker kafka producer
	svc.dataBakerProducer, err = getKafkaProducer(ctx, svc.cfg.Brokers, svc.cfg.DatabakerImportTopic, svc.cfg.KafkaMaxBytes)
	if err != nil {
		log.Event(ctx, "databaker kafka producer error", log.FATAL, log.Error(err))
		return err
	}

	// Get input file available kafka producer
	svc.inputFileAvailableProducer, err = getKafkaProducer(ctx, svc.cfg.Brokers, svc.cfg.InputFileAvailableTopic, svc.cfg.KafkaMaxBytes)
	if err != nil {
		log.Event(ctx, "direct kafka producer error", log.FATAL, log.Error(err))
		return err
	}

	// Create Identity Client
	svc.identityClient = clientsidentity.New(svc.cfg.ZebedeeURL)

	// Create dataset and recie API clients.
	// TODO: We should consider replacing these with the corresponding dp-api-clients-go clients
	client := dphttp.NewClient()
	svc.datasetAPI = &dataset.API{Client: client, URL: svc.cfg.DatasetAPIURL, ServiceAuthToken: svc.cfg.ServiceAuthToken}
	svc.recipeAPI = &recipe.API{Client: client, URL: svc.cfg.RecipeAPIURL}

	// Get HealthCheck and register checkers
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "error creating version info", log.FATAL, log.Error(err))
		return err
	}
	svc.healthCheck = getHealthCheck(versionInfo, svc.cfg.HealthCheckCriticalTimeout, svc.cfg.HealthCheckInterval)
	if err := svc.registerCheckers(ctx); err != nil {
		return errors.Wrap(err, "unable to register checkers")
	}

	// Get HTTP router and server with middleware
	r := mux.NewRouter()
	m := svc.createMiddleware(svc.cfg)
	svc.server = getHTTPServer(svc.cfg.BindAddr, m.Then(r))

	// Create API with job service
	urlBuilder := url.NewBuilder(svc.cfg.Host, svc.cfg.DatasetAPIURL)
	jobQueue := importqueue.CreateImportQueue(svc.dataBakerProducer.Channels().Output, svc.inputFileAvailableProducer.Channels().Output)
	jobService := job.NewService(svc.mongoDataStore, jobQueue, svc.datasetAPI, svc.recipeAPI, urlBuilder)
	svc.importAPI = api.Setup(r, svc.mongoDataStore, jobService, cfg)
	return nil
}

// Start starts an initialised service
func (svc *Service) Start(ctx context.Context, svcErrors chan error) {

	// Start kafka logging
	svc.dataBakerProducer.Channels().LogErrors(ctx, "error received from kafka data baker producer, topic: "+svc.cfg.DatabakerImportTopic)
	svc.inputFileAvailableProducer.Channels().LogErrors(ctx, "error received from kafka input file available producer, topic: "+svc.cfg.InputFileAvailableTopic)

	// Start healthcheck
	svc.healthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		log.Event(ctx, "Starting api...", log.INFO)
		if err := svc.server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()
}

// CreateMiddleware creates an Alice middleware chain of handlers
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
		if svc.healthCheck != nil {
			svc.healthCheck.Stop()
		}

		// stop any incoming requests
		if svc.server != nil {
			if err := svc.server.Shutdown(ctx); err != nil {
				log.Event(ctx, "failed to shutdown http server", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
		}

		// Close MongoDB (if it exists)
		if svc.mongoDataStore != nil {
			log.Event(ctx, "closing mongo data store", log.INFO)
			if err := svc.mongoDataStore.Close(ctx); err != nil {
				log.Event(ctx, "unable to close mongo data store", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
		}

		// Close Data Baker Kafka Producer (it if exists)
		if svc.dataBakerProducer != nil {
			log.Event(ctx, "closing data baker producer", log.INFO)
			if err := svc.dataBakerProducer.Close(ctx); err != nil {
				log.Event(ctx, "unable to close data baker producer", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
		}

		// Close Direct Kafka Producer (if it exists)
		if svc.inputFileAvailableProducer != nil {
			log.Event(ctx, "closing direct producer", log.INFO)
			if err := svc.inputFileAvailableProducer.Close(ctx); err != nil {
				log.Event(ctx, "unable to close direct producer", log.ERROR, log.Error(err))
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

// registerCheckers adds the checkers for the service clients to the health check object.
// If any dependency is missing, an erroring check will be created for it. This covers the scenario
// when a service is started with some missing dependency.
func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false

	// generic interface that must be satisfied by all health-checkable dependencies
	type Dependency interface {
		Checker(context.Context, *healthcheck.CheckState) error
	}

	// generic register checker method - if dependency is nil, a failing healthcheck will be created.
	registerChecker := func(name string, dependency Dependency) {
		if dependency != nil {
			if err = svc.healthCheck.AddCheck(name, dependency.Checker); err != nil {
				log.Event(ctx, fmt.Sprintf("error creating %s health check", strings.ToLower(name)), log.ERROR, log.Error(err))
				hasErrors = true
			}
		} else {
			svc.healthCheck.AddCheck(name, func(ctx context.Context, state *healthcheck.CheckState) error {
				err := errors.New(fmt.Sprintf("%s not initialised", strings.ToLower(name)))
				state.Update(healthcheck.StatusCritical, err.Error(), 0)
				return err
			})
		}
	}

	datasetAPIClient := clientshealth.NewClientWithClienter("dataset-api", svc.datasetAPI.URL, svc.datasetAPI.Client)
	recipeAPIHealthCheckClient := clientshealth.NewClientWithClienter("recipe-api", svc.recipeAPI.URL, svc.recipeAPI.Client)

	registerChecker("Kafka Data Baker Producer", svc.dataBakerProducer)
	registerChecker("Kafka Input File Available Producer", svc.inputFileAvailableProducer)
	registerChecker("Zebedee", svc.identityClient)
	registerChecker("Mongo DB", svc.mongoDataStore)
	registerChecker("Dataset API", datasetAPIClient)
	registerChecker("Recipe API", recipeAPIHealthCheckClient)

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
