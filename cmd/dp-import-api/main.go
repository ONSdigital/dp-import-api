package main

import (
	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/postgres"
	"github.com/ONSdigital/go-ns/log"

	"database/sql"
	"net/http"
	"os"

	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/importqueue"
	"github.com/ONSdigital/go-ns/kafka"
	_ "github.com/lib/pq"
)

func main() {
	log.Namespace = "dp-import-api"
	configuration, configErr := config.Get()
	if configErr != nil {
		log.Error(configErr, nil)
		os.Exit(1)
	}

	log.Info("Starting importqueue api", log.Data{"BIND_ADDR": configuration.BindAddr,
		"TOPICS":  []string{configuration.DatabakerImportTopic, configuration.InputFileAvailableTopic},
		"BROKERS": configuration.Brokers})
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
	dataBakerProducer := kafka.NewProducer(configuration.Brokers, configuration.DatabakerImportTopic, configuration.KafkaMaxBytes)
	directProducer := kafka.NewProducer(configuration.Brokers, configuration.InputFileAvailableTopic, configuration.KafkaMaxBytes)
	jobQueue := importqueue.CreateImportQueue(dataBakerProducer.Output(), directProducer.Output())
	importAPI := api.CreateImportAPI(configuration.Host, postgresDataStore, &jobQueue)
	httpCloseError := http.ListenAndServe(configuration.BindAddr, importAPI.Router)
	if httpCloseError != nil {
		log.Error(httpCloseError, log.Data{"BIND_ADDR": configuration.BindAddr,
			"TOPICS": []string{configuration.DatabakerImportTopic, configuration.InputFileAvailableTopic}})
	}
	dataBakerProducer.Closer() <- true
	directProducer.Closer() <- true
}
