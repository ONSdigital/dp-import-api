package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/middleware"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/importqueue"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/mongo"
	"github.com/ONSdigital/dp-import-api/url"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the Dataset API
type Service struct {
	cfg                                      *config.Configuration
	mongoDataStore                           datastore.DataStorer
	dataBakerProducer                        kafka.IProducer
	inputFileAvailableProducer               kafka.IProducer
	cantabularDatasetInstanceStartedProducer kafka.IProducer
	server                                   HTTPServer
	healthCheck                              HealthChecker
	importAPI                                *api.ImportAPI
	identityClient                           *clientsidentity.Client
	datasetAPIClient                         job.DatasetAPIClient
	recipeAPIClient                          job.RecipeAPIClient
}

// getMongoDataStore creates a mongoDB connection
var getMongoDataStore = func(ctx context.Context, cfg *config.Configuration) (datastore.DataStorer, error) {
	return mongo.NewDatastore(ctx, cfg.MongoDBURL, cfg.MongoDBDatabase, cfg.MongoDBCollection)
}

// getKafkaProducer creates a new Kafka Producer
var getKafkaProducer = func(ctx context.Context, cfg *config.Configuration, topic string) (kafka.IProducer, error) {
	producerChannels := kafka.CreateProducerChannels()
	pConfig := &kafka.ProducerConfig{
		KafkaVersion:    &cfg.KafkaVersion,
		MaxMessageBytes: &cfg.KafkaMaxBytes,
	}

	if cfg.KafkaSecProtocol == config.KafkaSecProtocolTLS {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaSecCACerts,
			cfg.KafkaSecClientCert,
			cfg.KafkaSecClientKey,
			cfg.KafkaSecSkipVerify,
		)
	}

	return kafka.NewProducer(ctx, cfg.KafkaAddr, topic, producerChannels, pConfig)
}

// getLegacyKafkaProducer creates a new Kafka Producer for legacy kafka
var getLegacyKafkaProducer = func(ctx context.Context, cfg *config.Configuration, topic string) (kafka.IProducer, error) {
	producerChannels := kafka.CreateProducerChannels()
	pConfig := &kafka.ProducerConfig{
		KafkaVersion:    &cfg.KafkaLegacyVersion,
		MaxMessageBytes: &cfg.KafkaMaxBytes,
	}

	return kafka.NewProducer(ctx, cfg.KafkaLegacyAddr, topic, producerChannels, pConfig)
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
	svc.mongoDataStore, err = getMongoDataStore(ctx, svc.cfg)
	if err != nil {
		log.Error(ctx, "mongodb datastore error", err)
	}

	// Get data baker kafka producer
	svc.dataBakerProducer, err = getKafkaProducer(ctx, svc.cfg, svc.cfg.DatabakerImportTopic)
	if err != nil {
		log.Fatal(ctx, "databaker kafka producer error", err)
		return err
	}

	// Get input file available kafka producer
	svc.inputFileAvailableProducer, err = getKafkaProducer(ctx, svc.cfg, svc.cfg.InputFileAvailableTopic)
	if err != nil {
		log.Fatal(ctx, "direct kafka producer error", err)
		return err
	}

	// Get Cantabular Dataset Instance Started kafka producer
	svc.cantabularDatasetInstanceStartedProducer, err = getLegacyKafkaProducer(ctx, svc.cfg, svc.cfg.CantabularDatasetInstanceStartedTopic)
	if err != nil {
		log.Fatal(ctx, "cantabular dataset instance started kafka producer error", err)
		return err
	}

	// Create Identity Client
	svc.identityClient = clientsidentity.New(svc.cfg.ZebedeeURL)

	// Create dataset and recipe API clients.
	svc.datasetAPIClient = dataset.NewAPIClient(svc.cfg.DatasetAPIURL)
	svc.recipeAPIClient = recipe.NewClient(svc.cfg.RecipeAPIURL)

	// Get HealthCheck and register checkers
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "error creating version info", err)
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
	jobQueue := importqueue.CreateImportQueue(
		svc.dataBakerProducer.Channels().Output,
		svc.inputFileAvailableProducer.Channels().Output,
		svc.cantabularDatasetInstanceStartedProducer.Channels().Output,
	)
	jobService := job.NewService(svc.mongoDataStore, jobQueue, svc.cfg.DatasetAPIURL, svc.datasetAPIClient, svc.recipeAPIClient, urlBuilder, svc.cfg.ServiceAuthToken)
	svc.importAPI = api.Setup(r, svc.mongoDataStore, jobService, cfg)
	return nil
}

// Start starts an initialised service
func (svc *Service) Start(ctx context.Context, svcErrors chan error) {

	// Start kafka logging
	svc.dataBakerProducer.Channels().LogErrors(ctx, "error received from kafka data baker producer, topic: "+svc.cfg.DatabakerImportTopic)
	svc.inputFileAvailableProducer.Channels().LogErrors(ctx, "error received from kafka input file available producer, topic: "+svc.cfg.InputFileAvailableTopic)
	svc.cantabularDatasetInstanceStartedProducer.Channels().LogErrors(ctx, "error recieved from kafka cantabular dataset instance started producer, topic: "+svc.cfg.CantabularDatasetInstanceStartedTopic)

	// Start healthcheck
	svc.healthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		log.Info(ctx, "Starting api...")
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
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
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
				log.Error(ctx, "failed to shutdown http server", err)
				hasShutdownError = true
			}
		}

		// Close MongoDB (if it exists)
		if svc.mongoDataStore != nil {
			log.Info(ctx, "closing mongo data store")
			if err := svc.mongoDataStore.Close(ctx); err != nil {
				log.Error(ctx, "unable to close mongo data store", err)
				hasShutdownError = true
			}
		}

		// Close Data Baker Kafka Producer (it if exists)
		if svc.dataBakerProducer != nil {
			log.Info(ctx, "closing data baker producer")
			if err := svc.dataBakerProducer.Close(ctx); err != nil {
				log.Error(ctx, "unable to close data baker producer", err)
				hasShutdownError = true
			}
		}

		// Close Direct Kafka Producer (if it exists)
		if svc.inputFileAvailableProducer != nil {
			log.Info(ctx, "closing direct producer")
			if err := svc.inputFileAvailableProducer.Close(ctx); err != nil {
				log.Error(ctx, "unable to close direct producer", err)
				hasShutdownError = true
			}
		}

		// Close Cantabular Kafka Producer (if it exists)
		if svc.cantabularDatasetInstanceStartedProducer != nil {
			log.Info(ctx, "closing cantabular dataset instance started producer")
			if err := svc.cantabularDatasetInstanceStartedProducer.Close(ctx); err != nil {
				log.Error(ctx, "unable to close cantabular dataset instance started producer", err)
				hasShutdownError = true
			}
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "shutdown timed out", ctx.Err())
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
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
				log.Error(ctx, fmt.Sprintf("error creating %s health check", strings.ToLower(name)), err)
				hasErrors = true
			}
		} else {
			if err = svc.healthCheck.AddCheck(name, func(ctx context.Context, state *healthcheck.CheckState) error {
				message := fmt.Sprintf("%s not initialised", strings.ToLower(name))
				return state.Update(healthcheck.StatusCritical, message, 0)
			}); err != nil {
				log.Error(ctx, fmt.Sprintf("error creating %s health check stub for unused (nil) dependency", strings.ToLower(name)), err)
				hasErrors = true
			}
		}
	}

	registerChecker("Kafka Data Baker Producer", svc.dataBakerProducer)
	registerChecker("Kafka Input File Available Producer", svc.inputFileAvailableProducer)
	registerChecker("Kafka Cantabular Dataset Instance Started Producer", svc.cantabularDatasetInstanceStartedProducer)
	registerChecker("Zebedee", svc.identityClient)
	registerChecker("Mongo DB", svc.mongoDataStore)
	registerChecker("Dataset API", svc.datasetAPIClient)
	registerChecker("Recipe API", svc.recipeAPIClient)

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
