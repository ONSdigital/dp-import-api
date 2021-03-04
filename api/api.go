package api

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-import-api/config"
	"net/http"

	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-net/handlers"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

//go:generate moq -out testapi/job_service.go -pkg testapi . JobService

const (
	jobIDKey = "job_id"
)

// ImportAPI is a restful API used to manage importing datasets to be published
type ImportAPI struct {
	dataStore     datastore.DataStorer
	router        *mux.Router
	jobService    JobService
	defaultLimit  int
	defaultOffset int
	maxLimit      int
}

// JobService provide business logic for job related operations.
type JobService interface {
	CreateJob(ctx context.Context, job *models.Job) (*models.Job, error)
	UpdateJob(ctx context.Context, jobID string, job *models.Job) error
}

// Setup manages all the routes configured to API
func Setup(router *mux.Router,
	dataStore datastore.DataStorer,
	jobService JobService, cfg *config.Configuration) *ImportAPI {

	api := &ImportAPI{
		dataStore:     dataStore,
		router:        router,
		jobService:    jobService,
		defaultLimit:  cfg.DefaultLimit,
		defaultOffset: cfg.DefaultOffset,
		maxLimit:      cfg.DefaultMaxLimit,
	}

	// External API for florence
	api.router.Path("/jobs").Methods("POST").HandlerFunc(handlers.CheckIdentity(api.addJobHandler))
	api.router.Path("/jobs").Methods("GET").HandlerFunc(handlers.CheckIdentity(api.getJobsHandler))
	api.router.Path("/jobs/{id}").Methods("GET").HandlerFunc(handlers.CheckIdentity(api.getJobHandler))
	api.router.Path("/jobs/{id}").Methods("PUT").HandlerFunc(handlers.CheckIdentity(api.updateJobHandler))
	api.router.Path("/jobs/{id}/files").Methods("PUT").HandlerFunc(handlers.CheckIdentity(api.addUploadedFileHandler))
	return api
}

func writeResponse(ctx context.Context, w http.ResponseWriter, statusCode int, b []byte, action string, logData log.Data) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(b); err != nil {
		log.Event(ctx, fmt.Sprintf("%s endpoint: failed to write response body", action), log.ERROR, log.Error(err), logData)
	}
}

func handleErr(ctx context.Context, w http.ResponseWriter, err error, logData log.Data) {
	if logData == nil {
		logData = log.Data{}
	}

	var status int
	response := err

	switch {
	case errs.NotFoundMap[err]:
		status = http.StatusNotFound
	case errs.BadRequestMap[err]:
		status = http.StatusBadRequest
	default:
		status = http.StatusInternalServerError
		response = errs.ErrInternalServer
	}

	logData["responseStatus"] = status
	log.Event(ctx, "request unsuccessful", log.ERROR, log.Error(err), logData)
	http.Error(w, response.Error(), status)
}
