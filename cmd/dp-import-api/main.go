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
	"github.com/ONSdigital/dp-import-api/event"
	"github.com/ONSdigital/dp-import-api/importqueue"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/mongo"
	"github.com/ONSdigital/dp-import-api/recipe"
	"github.com/ONSdigital/dp-import-api/url"
	"github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/ONSdigital/go-ns/server"
	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
)

func main() {
	log.Namespace = "dp-import-api"
	config, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
	client := rchttp.DefaultClient

	log.Info("Starting import api", log.Data{
		"bind_addr": config.BindAddr,
		"topics":    []string{config.DatabakerImportTopic, config.InputFileAvailableTopic},
		"brokers":   config.Brokers,
	})

	mongoDataStore, err := mongo.NewDatastore(config.MongoDBURL, config.MongoDBDatabase, config.MongoDBCollection)
	if err != nil {
		log.ErrorC("mongodb datastore error", err, nil)
		os.Exit(1)
	}

	dataBakerProducer, err := kafka.NewProducer(config.Brokers, config.DatabakerImportTopic, config.KafkaMaxBytes)
	if err != nil {
		log.ErrorC("data baker kafka producer error", err, nil)
		os.Exit(1)
	}

	directProducer, err := kafka.NewProducer(config.Brokers, config.InputFileAvailableTopic, config.KafkaMaxBytes)
	if err != nil {
		log.ErrorC("direct kafka producer error", err, nil)
		os.Exit(1)
	}

	router := mux.NewRouter()
	router.Path("/healthcheck").HandlerFunc(healthcheck.Handler)

	httpServer := server.New(config.BindAddr, router)
	httpServer.HandleOSSignals = false

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	urlBuilder := url.NewBuilder(config.Host, config.DatasetAPIURL)
	jobQueue := importqueue.CreateImportQueue(dataBakerProducer.Output(), directProducer.Output())

	datasetAPI := dataset.API{client, config.DatasetAPIURL, config.DatasetAPIAuthToken}
	recipeAPI := recipe.API{client, config.RecipeAPIURL}

	jobService := job.NewService(mongoDataStore, jobQueue, &datasetAPI, &recipeAPI, urlBuilder)
	_ = api.CreateImportAPI(router, mongoDataStore, config.SecretKey, jobService)

	// signals the web server shutdown, so a graceful exit is required
	httpErrChannel := make(chan error)
	// launch web server in background
	go func() {
		log.Debug("listening...", log.Data{"bind_address": config.BindAddr})
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error(err, nil)
			httpErrChannel <- err
			return
		}
		httpErrChannel <- errors.New("http server completed - with no error")
	}()

	observationsImportedConsumer, err := kafka.NewConsumerGroup(config.Brokers, config.ObservationsImportedTopic, log.Namespace, sarama.OffsetOldest)
	if err != nil {
		log.ErrorC("error creating kafka consumer", err, nil)
		os.Exit(1)
	}

	observationsImportedHandler := event.NewObservationsImportedConsumer()
	observationsImportedHandler.Consume(observationsImportedConsumer, jobService)

	shutdownGracefully := func(httpDead bool) {
		// gracefully retire resources
		ctx, cancel := context.WithTimeout(context.Background(), config.GracefulShutdownTimeout)
		defer cancel()

		if !httpDead {
			if err = httpServer.Shutdown(ctx); err != nil {
				log.Error(err, nil)
			}
		}

		if err = dataBakerProducer.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		if err = directProducer.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		if err = observationsImportedHandler.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		if err = observationsImportedConsumer.Close(ctx); err != nil {
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
	case err := <-httpErrChannel:
		log.ErrorC("error channel", err, nil)
		shutdownGracefully(true)
	case sig := <-signals:
		log.Error(errors.New("os signal received"), log.Data{"signal": sig.String()})
		shutdownGracefully(false)
	}
}