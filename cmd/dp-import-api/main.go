package main

import (
	"github.com/ONSdigital/dp-import-api/api"
	"github.com/ONSdigital/dp-import-api/postgres"
	"github.com/ONSdigital/dp-import-api/utils"
	"github.com/ONSdigital/go-ns/log"

	"database/sql"
	_ "github.com/lib/pq"
	"net/http"
)

func main() {
	log.Namespace = "dp-dp-import-api"
	address := utils.GetEnvVariable("BIND_ADDR", ":21800")
	dbUrl := utils.GetEnvVariable("DB_URL", "user=dp dbname=ImportJobs sslmode=disable")
	log.Debug("Starting import api", log.Data{"BIND_ADDR": address})
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.ErrorC("DB open error", err, nil)
		panic("DB open error")
	}
	// LIMIT CONNECTIONS HERRE!!!!!!!!
	postgresDataStore, error := postgres.NewDatastore(db)
	if error != nil {
		log.ErrorC("Create postgres error", err, nil)
		panic("Create data Store error")
	}
	importAPI := api.CreateImportAPI(postgresDataStore)
	httpCloseError := http.ListenAndServe(address, importAPI.Router)
	if httpCloseError != nil {
		log.Error(httpCloseError, log.Data{"BIND_ADDR": address})
	}
}
