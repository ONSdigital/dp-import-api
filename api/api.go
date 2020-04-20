package api

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/middleware"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/identity"
	"net/http"

	identityclient "github.com/ONSdigital/dp-api-clients-go/identity"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/server"
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

var httpServer *server.Server

// ImportAPI is a restful API used to manage importing datasets to be published
type ImportAPI struct {
	dataStore  datastore.DataStorer
	router     *mux.Router
	jobService JobService
	auditor    Auditor
}

// JobService provide business logic for job related operations.
type JobService interface {
	CreateJob(ctx context.Context, job *models.Job) (*models.Job, error)
	UpdateJob(ctx context.Context, jobID string, job *models.Job) error
}

// Auditor provides auditor service
type Auditor audit.AuditorService

// CreateImportAPI manages all the routes configured to API
func CreateImportAPI(ctx context.Context,
	bindAddr, zebedeeURL string,
	mongoDataStore datastore.DataStorer,
	jobService JobService,
	auditor audit.AuditorService,
	hc *healthcheck.HealthCheck) {

	router := mux.NewRouter()
	routes(router, mongoDataStore, jobService, auditor, hc)

	identityHTTPClient := rchttp.NewClient()
	identityClient := identityclient.NewAPIClient(identityHTTPClient, zebedeeURL)
	identityHandler := identity.HandlerForHTTPClient(identityClient)

	middleware := alice.New(
		middleware.Whitelist(middleware.HealthcheckFilter(hc.Handler)),
		requestID.Handler(16),
		identityHandler,
	).Then(router)

	httpServer = server.New(bindAddr, middleware)
	httpServer.HandleOSSignals = false

	go func() {
		log.Event(ctx, "Starting api...", log.INFO)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Event(ctx, "api http server returned error", log.ERROR, log.Error(err))
		}
	}()
}

// routes contain all endpoints for API
func routes(router *mux.Router, dataStore datastore.DataStorer, jobService JobService, auditor Auditor, hc *healthcheck.HealthCheck) *ImportAPI {
	api := ImportAPI{dataStore: dataStore, router: router, jobService: jobService, auditor: auditor}

	// External API for florence
	api.router.Path("/jobs").Methods("POST").HandlerFunc(identity.Check(auditor, addJobAction, api.addJobHandler))
	api.router.Path("/jobs").Methods("GET").HandlerFunc(identity.Check(auditor, getJobsAction, api.getJobsHandler))
	api.router.Path("/jobs/{id}").Methods("GET").HandlerFunc(identity.Check(auditor, getJobAction, api.getJobHandler))
	api.router.Path("/jobs/{id}").Methods("PUT").HandlerFunc(identity.Check(auditor, updateJobAction, api.updateJobHandler))
	api.router.Path("/jobs/{id}/files").Methods("PUT").HandlerFunc(identity.Check(auditor, uploadFileAction, api.addUploadedFileHandler))
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
