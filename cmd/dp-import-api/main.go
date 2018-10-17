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
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/mongo"
	"github.com/ONSdigital/dp-import-api/recipe"
	"github.com/ONSdigital/dp-import-api/url"
	"github.com/ONSdigital/go-ns/audit"
	handlershealthcheck "github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

const serviceNamespace = "dp-import-api"

func main() {
	log.Namespace = serviceNamespace
	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
	client := rchttp.NewClient()

	// sensitive fields are omitted from config.String().
	log.Info("loaded config", log.Data{
		"config": cfg,
	})

	mongoDataStore, err := mongo.NewDatastore(cfg.MongoDBURL, cfg.MongoDBDatabase, cfg.MongoDBCollection)
	if err != nil {
		log.ErrorC("mongodb datastore error", err, nil)
		os.Exit(1)
	}

	dataBakerProducer, err := kafka.NewProducer(cfg.Brokers, cfg.DatabakerImportTopic, cfg.KafkaMaxBytes)
	if err != nil {
		log.ErrorC("data baker kafka producer error", err, nil)
		os.Exit(1)
	}

	directProducer, err := kafka.NewProducer(cfg.Brokers, cfg.InputFileAvailableTopic, cfg.KafkaMaxBytes)
	if err != nil {
		log.ErrorC("direct kafka producer error", err, nil)
		os.Exit(1)
	}
	auditProducer, err := kafka.NewProducer(cfg.Brokers, cfg.AuditEventsTopic, cfg.KafkaMaxBytes)
	if err != nil {
		log.ErrorC("direct kafka producer error", err, nil)
		os.Exit(1)
	}

	router := mux.NewRouter()

	healthcheckHandler := healthcheck.NewMiddleware(handlershealthcheck.Handler)
	identityHandler := identity.Handler(cfg.ZebedeeURL)

	// TODO how long should the ID be?
	alice := alice.New(requestID.Handler(16), healthcheckHandler, identityHandler).Then(router)

	httpServer := server.New(cfg.BindAddr, alice)
	httpServer.HandleOSSignals = false

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	urlBuilder := url.NewBuilder(cfg.Host, cfg.DatasetAPIURL)
	jobQueue := importqueue.CreateImportQueue(dataBakerProducer.Output(), directProducer.Output())

	datasetAPI := dataset.API{client, cfg.DatasetAPIURL, cfg.ServiceAuthToken}
	recipeAPI := recipe.API{client, cfg.RecipeAPIURL}

	jobService := job.NewService(mongoDataStore, jobQueue, &datasetAPI, &recipeAPI, urlBuilder)
	auditor := audit.New(auditProducer, serviceNamespace)

	_ = api.CreateImportAPI(router, mongoDataStore, jobService, auditor)

	// signals the web server shutdown, so a graceful exit is required
	httpErrChannel := make(chan error)
	// launch web server in background
	go func() {
		log.Debug("listening...", log.Data{"bind_address": cfg.BindAddr})
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error(err, nil)
			httpErrChannel <- err
			return
		}
		httpErrChannel <- errors.New("http server completed - with no error")
	}()

	shutdownGracefully := func(httpDead bool) {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
		defer cancel()

		if !httpDead {
			if err = httpServer.Shutdown(ctx); err != nil {
				log.Error(err, nil)
			}
		}

		if err = auditProducer.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		if err = dataBakerProducer.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		if err = directProducer.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		// mongo.Close() may use all remaining time in the context - do this last!
		if err = mongoDataStore.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		log.Debug("graceful shutdown has completed", nil)
		os.Exit(1)
	}

	select {
	case err := <-dataBakerProducer.Errors():
		log.ErrorC("kafka databaker producer", err, nil)
		shutdownGracefully(false)
	case err := <-directProducer.Errors():
		log.ErrorC("kafka direct producer", err, nil)
		shutdownGracefully(false)
	case err := <-auditProducer.Errors():
		log.ErrorC("kafka audit producer", err, nil)
		shutdownGracefully(false)
	case err := <-httpErrChannel:
		log.ErrorC("error channel", err, nil)
		shutdownGracefully(true)
	case sig := <-signals:
		log.Error(errors.New("os signal received"), log.Data{"signal": sig.String()})
		shutdownGracefully(false)
	}

}
