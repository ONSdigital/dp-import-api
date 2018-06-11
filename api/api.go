package api

import (
	"context"
	"net/http"

	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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

// CreateImportAPI returns the api with all the routes configured
func CreateImportAPI(router *mux.Router, dataStore datastore.DataStorer, jobService JobService, auditor Auditor) *ImportAPI {

	api := ImportAPI{dataStore: dataStore, router: router, jobService: jobService, auditor: auditor}

	// External API for florence
	api.router.Path("/jobs").Methods("POST").HandlerFunc(identity.Check(api.addJobHandler))
	api.router.Path("/jobs").Methods("GET").HandlerFunc(identity.Check(api.getJobsHandler)).Queries()
	api.router.Path("/jobs/{id}").Methods("GET").HandlerFunc(identity.Check(api.getJobHandler))
	api.router.Path("/jobs/{id}").Methods("PUT").HandlerFunc(identity.Check(api.updateJobHandler))
	api.router.Path("/jobs/{id}/files").Methods("PUT").HandlerFunc(identity.Check(api.addUploadedFileHandler))
	api.router.NotFoundHandler = &api
	return &api
}

func (api *ImportAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, notFoundError, http.StatusNotFound)
}

func writeBody(ctx context.Context, w http.ResponseWriter, b []byte, action string, data log.Data) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		log.ErrorCtx(ctx, errors.Wrapf(err, "%s endpoint: failed to write response body", action), data)
	}
}

func handleErr(ctx context.Context, w http.ResponseWriter, err error, data log.Data) {
	if data == nil {
		data = log.Data{}
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

	data["responseStatus"] = status
	audit.LogError(ctx, errors.WithMessage(err, "request unsuccessful"), data)

	http.Error(w, response.Error(), status)
}
