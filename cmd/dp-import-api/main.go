package main

import (
	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/postgres"
	"github.com/ONSdigital/go-ns/log"

	"database/sql"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/importqueue"
	"github.com/ONSdigital/go-ns/kafka"
	_ "github.com/lib/pq"
)

func main() {
	log.Namespace = "dp-import-api"
	config, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	log.Info("Starting importqueue api", log.Data{"bind_addr": config.BindAddr,
		"topics":  []string{config.DatabakerImportTopic, config.InputFileAvailableTopic},
		"brokers": config.Brokers})
	db, err := sql.Open("postgres", config.PostgresURL)
	if err != nil {
		log.ErrorC("DB open error", err, nil)
		os.Exit(1)
	}
	postgresDataStore, err := postgres.NewDatastore(db)
	if err != nil {
		log.ErrorC("Create postgres error", err, nil)
		os.Exit(1)
	}
	dataBakerProducer := kafka.NewProducer(config.Brokers, config.DatabakerImportTopic, config.KafkaMaxBytes)
	directProducer := kafka.NewProducer(config.Brokers, config.InputFileAvailableTopic, config.KafkaMaxBytes)

	jobQueue := importqueue.CreateImportQueue(dataBakerProducer.Output(), directProducer.Output())
	router := mux.NewRouter()
	_ = api.CreateImportAPI(config.Host,router, postgresDataStore, &jobQueue)
	err = http.ListenAndServe(config.BindAddr, router)

	if err != nil {
		log.Error(err, log.Data{"bind_addr": config.BindAddr,
			"topic": []string{config.DatabakerImportTopic, config.InputFileAvailableTopic}})
	}
	dataBakerProducer.Closer() <- true
	directProducer.Closer() <- true
}
