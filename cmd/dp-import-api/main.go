package main

import (
	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/dataset"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"

	"os"

	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/importqueue"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/mongo"
	"github.com/ONSdigital/dp-import-api/recipe"
	"github.com/ONSdigital/dp-import-api/url"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/rhttp"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	log.Namespace = "dp-import-api"
	config, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
	client := rhttp.DefaultClient

	log.Info("Starting importqueue api", log.Data{
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
	s := server.New(config.BindAddr, router)
	log.Debug("listening...", log.Data{
		"bind_address": config.BindAddr,
	})

	jobQueue := importqueue.CreateImportQueue(dataBakerProducer.Output(), directProducer.Output())
	datasetAPI := dataset.NewDatasetAPI(client, config.DatasetAPIURL, config.DatasetAPIAuthToken)
	recipeAPI := recipe.NewAPI(client)
	urlBuilder := url.NewBuilder(config.Host, config.DatasetAPIURL)

	jobService := job.NewService(mongoDataStore, jobQueue, datasetAPI, recipeAPI, urlBuilder)

	_ = api.CreateImportAPI(router, mongoDataStore, config.SecretKey, jobService)
	if err = s.ListenAndServe(); err != nil {
		log.Error(err, nil)
	}

	dataBakerProducer.Closer() <- true
	directProducer.Closer() <- true
}
