package api

import (
	"context"
	"fmt"
	"net/http"

	identityclient "github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/dp-api-clients-go/middleware"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-net/handlers"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

//go:generate moq -out testapi/job_service.go -pkg testapi . JobService

const (
	// audit actions
	uploadFileAction = "uploadFile"
	updateJobAction  = "updateJob"
	addJobAction     = "addJob"
	getJobAction     = "getJob"
	getJobsAction    = "getJobs"

	jobIDKey      = "job_id"
	notFoundError = "requested resource not found"
)

var httpServer *dphttp.Server

// ImportAPI is a restful API used to manage importing datasets to be published
type ImportAPI struct {
	dataStore  datastore.DataStorer
	router     *mux.Router
	jobService JobService
}

// JobService provide business logic for job related operations.
type JobService interface {
	CreateJob(ctx context.Context, job *models.Job) (*models.Job, error)
	UpdateJob(ctx context.Context, jobID string, job *models.Job) error
}

// CreateImportAPI manages all the routes configured to API
func CreateImportAPI(ctx context.Context,
	bindAddr, zebedeeURL string,
	mongoDataStore datastore.DataStorer,
	jobService JobService,
	hc *healthcheck.HealthCheck) {

	router := mux.NewRouter()
	routes(router, mongoDataStore, jobService, hc)

	identityClient := identityclient.New(zebedeeURL)
	identityHandler := handlers.IdentityWithHTTPClient(identityClient)

	middleware := alice.New(
		middleware.Whitelist(middleware.HealthcheckFilter(hc.Handler)),
		dprequest.HandlerRequestID(16),
		identityHandler,
	).Then(router)

	httpServer = dphttp.NewServer(bindAddr, middleware)
	httpServer.HandleOSSignals = false

	go func() {
		log.Event(ctx, "Starting api...", log.INFO)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Event(ctx, "api http server returned error", log.ERROR, log.Error(err))
		}
	}()
}

// routes contain all endpoints for API
func routes(router *mux.Router, dataStore datastore.DataStorer, jobService JobService, hc *healthcheck.HealthCheck) *ImportAPI {
	api := ImportAPI{dataStore: dataStore, router: router, jobService: jobService}

	// External API for florence
	api.router.Path("/jobs").Methods("POST").HandlerFunc(handlers.CheckIdentity(api.addJobHandler))
	api.router.Path("/jobs").Methods("GET").HandlerFunc(handlers.CheckIdentity(api.getJobsHandler))
	api.router.Path("/jobs/{id}").Methods("GET").HandlerFunc(handlers.CheckIdentity(api.getJobHandler))
	api.router.Path("/jobs/{id}").Methods("PUT").HandlerFunc(handlers.CheckIdentity(api.updateJobHandler))
	api.router.Path("/jobs/{id}/files").Methods("PUT").HandlerFunc(handlers.CheckIdentity(api.addUploadedFileHandler))
	api.router.NotFoundHandler = &api
	return &api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Event(ctx, "graceful shutdown of http server complete", log.INFO)
	return nil
}

func (api *ImportAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, notFoundError, http.StatusNotFound)
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
