package main

import (
	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/postgres"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"

	"database/sql"
	"os"

	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/importqueue"
	"github.com/ONSdigital/go-ns/kafka"
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
		log.ErrorC("postgres datastore error", err, nil)
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

	jobQueue := importqueue.CreateImportQueue(dataBakerProducer.Output(), directProducer.Output())
	router := mux.NewRouter()

	s := server.New(config.BindAddr, router)

	log.Debug("listening...", log.Data{
		"bind_address": config.BindAddr,
	})

	_ = api.CreateImportAPI(config.Host, router, postgresDataStore, &jobQueue, config.SecretKey)
	if err = s.ListenAndServe(); err != nil {
		log.Error(err, nil)
	}

	dataBakerProducer.Closer() <- true
	directProducer.Closer() <- true
}
