package main

import (
	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/postgres"
	"github.com/ONSdigital/go-ns/log"

	"database/sql"
	_ "github.com/lib/pq"
	"net/http"
	"github.com/ONSdigital/dp-import-api/config"
	"os"
	"github.com/ONSdigital/go-ns/kafka"
)

func main() {
	log.Namespace = "dp-import-api"
	configuration, configErr := config.Get()
	if configErr != nil {
		log.Error(configErr, nil)
		os.Exit(1)
	}

	log.Debug("Starting import api", log.Data{"BIND_ADDR": configuration.BindAddr})
	db, postgresErr := sql.Open("postgres", configuration.PostgresURL)
	if postgresErr != nil {
		log.ErrorC("DB open error", postgresErr, nil)
		os.Exit(1)
	}
	postgresDataStore, dataStoreError := postgres.NewDatastore(db)
	if dataStoreError != nil {
		log.ErrorC("Create postgres error", dataStoreError, nil)
		os.Exit(1)
	}
	producer := kafka.NewProducer(configuration.Brokers, configuration.PublishDatasetTopic, configuration.KafkaMaxBytes)
	importAPI := api.CreateImportAPI(postgresDataStore, producer.Output)
	httpCloseError := http.ListenAndServe(configuration.BindAddr, importAPI.Router)
	if httpCloseError != nil {
		log.Error(httpCloseError, log.Data{"BIND_ADDR": configuration.BindAddr, "TOPIC": configuration.PublishDatasetTopic})
	}
	producer.Closer <- true
}
