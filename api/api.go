package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

//go:generate moq -out testapi/job_service.go -pkg testapi . JobService

const internalError = "Internal server error"

const notFoundError = "requested resource not found"

const (
	// audit actions
	uploadFileAction = "uploadFile"
	updateJobAction  = "updateJob"
	addJobAction     = "addJob"
	getJobAction     = "getJob"
	getJobsAction    = "getJobs"

	// audit results
	actionSuccessful    = "actionSuccess"
	actionUnsuccessful  = "actionUnsuccessful"
	actionAttempted     = "actionAttempted"
	auditError          = "error auditing action, failing request"
	jobIDKey            = "job_id"
	internalServerError = "Internal server error"

	auditActionErr = "failed to audit action"
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
	api.router.Path("/jobs").Methods("POST").HandlerFunc(identity.Check(api.addJob))
	api.router.Path("/jobs").Methods("GET").HandlerFunc(identity.Check(api.getJobs)).Queries()
	api.router.Path("/jobs/{id}").Methods("GET").HandlerFunc(identity.Check(api.getJob))
	api.router.Path("/jobs/{id}").Methods("PUT").HandlerFunc(identity.Check(api.updateJob))
	api.router.Path("/jobs/{id}/files").Methods("PUT").HandlerFunc(identity.Check(api.addUploadedFile))
	api.router.NotFoundHandler = &api
	return &api
}

func (api *ImportAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, notFoundError, http.StatusNotFound)
}

func (api *ImportAPI) addJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	job, err := models.CreateJob(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "api endpoint addJob error - Bad client request received"), nil)
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}

	logData := log.Data{"recipeID": job.RecipeID}
	auditParams := common.Params{"recipeID": job.RecipeID}
	if auditError := api.auditor.Record(ctx, addJobAction, actionAttempted, auditParams); auditError != nil {
		handleAuditError(ctx, w, addJobAction, actionAttempted, auditError, logData)
		return
	}

	var createdJob *models.Job
	b, err := func() (bytes []byte, err error) {
		createdJob, err = api.jobService.CreateJob(ctx, job)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "addJob endpoint: error creating job resource"), logData)
			return nil, err
		}

		logData["job"] = createdJob
		auditParams["createdJobID"] = createdJob.ID

		bytes, err = json.Marshal(createdJob)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "addJob endpoint: failed to marshal job resource into bytes"), logData)
			return nil, err
		}

		return bytes, nil
	}()

	if err != nil {
		log.Info("errored", nil)
		if auditError := api.auditor.Record(ctx, addJobAction, actionUnsuccessful, auditParams); auditError != nil {
			handleAuditError(ctx, w, addJobAction, actionUnsuccessful, auditError, logData)
			return
		}

		setErrorCode(w, err)
		return
	}

	if auditError := api.auditor.Record(ctx, addJobAction, actionSuccessful, auditParams); auditError != nil {
		auditActionFailure(ctx, addJobAction, actionSuccessful, auditError, logData)
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(b); err != nil {
		audit.LogError(ctx, errors.WithMessage(err, "addJob endpoint: error writing bytes to response"), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	audit.LogInfo(ctx, "created new import job", logData)
}

func (api *ImportAPI) getJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auditParams := common.Params{}
	logData := log.Data{}

	filtersQuery := r.URL.Query().Get("state")
	var filterList []string
	if filtersQuery == "" {
		filterList = nil
	} else {
		filterList = strings.Split(filtersQuery, ",")
		logData["filterQuery"] = filtersQuery
		auditParams["filterQuery"] = filtersQuery
	}

	if auditError := api.auditor.Record(ctx, getJobsAction, actionAttempted, auditParams); auditError != nil {
		handleAuditError(ctx, w, getJobsAction, actionAttempted, auditError, logData)
		return
	}

	b, err := func() (bytes []byte, err error) {
		jobs, err := api.dataStore.GetJobs(filterList)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "getJobs endpoint: failed to retrieve a list of jobs"), logData)
			return nil, err
		}
		logData["Jobs"] = jobs

		bytes, err = json.Marshal(jobs)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "getJobs endpoint: failed to marshal jobs resource into bytes"), logData)
			return nil, err
		}

		return bytes, nil
	}()

	setJSONContentType(w)

	if err != nil {
		if auditError := api.auditor.Record(ctx, getJobsAction, actionUnsuccessful, auditParams); auditError != nil {
			handleAuditError(ctx, w, getJobsAction, actionUnsuccessful, auditError, logData)
			return
		}

		setErrorCode(w, err)
		return
	}

	if auditError := api.auditor.Record(ctx, getJobsAction, actionSuccessful, auditParams); auditError != nil {
		handleAuditError(ctx, w, getJobsAction, actionSuccessful, auditError, logData)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		audit.LogError(ctx, errors.WithMessage(err, "getJobs endpoint: error writing bytes to response"), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	audit.LogInfo(ctx, "getJobs endpoint: request successful", logData)
}

func (api *ImportAPI) getJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}
	auditParams := common.Params{jobIDKey: jobID}
	ctx := r.Context()

	if auditError := api.auditor.Record(ctx, getJobAction, actionAttempted, auditParams); auditError != nil {
		handleAuditError(ctx, w, getJobAction, actionAttempted, auditError, logData)
		return
	}

	var job *models.Job
	b, err := func() (bytes []byte, err error) {
		job, err = api.dataStore.GetJob(jobID)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "getJob endpoint: failed to find job"), logData)
			return nil, err
		}

		logData["job"] = job

		bytes, err = json.Marshal(job)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "getJob endpoint: failed to marshal jobs resource into bytes"), logData)
			return nil, err
		}
		return bytes, nil
	}()

	setJSONContentType(w)

	if err != nil {
		if auditError := api.auditor.Record(ctx, getJobAction, actionUnsuccessful, auditParams); auditError != nil {
			handleAuditError(ctx, w, getJobAction, actionUnsuccessful, auditError, logData)
			return
		}

		setErrorCode(w, err)
		return
	}

	if auditError := api.auditor.Record(ctx, getJobAction, actionSuccessful, auditParams); auditError != nil {
		handleAuditError(ctx, w, getJobAction, actionSuccessful, auditError, logData)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		audit.LogError(ctx, errors.WithMessage(err, "getJob endpoint: error writing bytes to response"), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	audit.LogInfo(ctx, "getJob endpoint: request successful", logData)
}

func (api *ImportAPI) addUploadedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}
	auditParams := common.Params{jobIDKey: jobID}
	ctx := r.Context()

	if auditError := api.auditor.Record(ctx, uploadFileAction, actionAttempted, auditParams); auditError != nil {
		handleAuditError(ctx, w, uploadFileAction, actionAttempted, auditError, logData)
		return
	}

	if err := func() (err error) {
		uploadedFile, err := models.CreateUploadedFile(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "addUploadFile endpoint: failed to create uploaded file resource"), logData)
			return err
		}

		logData["file"] = uploadedFile
		auditParams["fileAlias"] = uploadedFile.AliasName
		auditParams["fileURL"] = uploadedFile.URL

		if err := api.dataStore.AddUploadedFile(jobID, uploadedFile); err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "addUploadFile endpoint: failed to store uploaded file resource"), logData)
			return err
		}

		return nil
	}(); err != nil {
		if auditError := api.auditor.Record(ctx, uploadFileAction, actionUnsuccessful, auditParams); auditError != nil {
			handleAuditError(ctx, w, uploadFileAction, actionUnsuccessful, auditError, logData)
			return
		}

		setErrorCode(w, err)
		return
	}

	if auditError := api.auditor.Record(ctx, uploadFileAction, actionSuccessful, auditParams); auditError != nil {
		auditActionFailure(ctx, uploadFileAction, actionSuccessful, auditError, logData)
		return
	}

	audit.LogInfo(ctx, "added uploaded file to job", logData)
}

func (api *ImportAPI) updateJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}
	auditParams := common.Params{jobIDKey: jobID}

	if auditError := api.auditor.Record(ctx, updateJobAction, actionAttempted, auditParams); auditError != nil {
		handleAuditError(ctx, w, updateJobAction, actionAttempted, auditError, logData)
		return
	}

	if err := func() (err error) {
		job, err := models.CreateJob(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "updateJob endpoint: failed to update job resource"), logData)
			return err
		}

		logData["job"] = job
		if err := api.jobService.UpdateJob(ctx, jobID, job); err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "updateJob endpoint: failed to store updated job resource"), logData)
			return err
		}

		return nil
	}(); err != nil {
		if auditError := api.auditor.Record(ctx, updateJobAction, actionUnsuccessful, auditParams); auditError != nil {
			handleAuditError(ctx, w, updateJobAction, actionUnsuccessful, auditError, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	if auditError := api.auditor.Record(ctx, updateJobAction, actionSuccessful, auditParams); auditError != nil {
		auditActionFailure(ctx, updateJobAction, actionSuccessful, auditError, logData)
		return
	}

	audit.LogInfo(ctx, "job update completed successfully", logData)
}

func setErrorCode(w http.ResponseWriter, err error) {
	switch {
	case err == api_errors.JobNotFoundError:
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	case err == job.ErrInvalidJob:
		http.Error(w, "the given job model is not valid", http.StatusBadRequest)
		return
	case err == job.ErrGetRecipeFailed:
		http.Error(w, "failed to get recipe data", http.StatusInternalServerError)
		return
	case err == job.ErrSaveJobFailed:
		http.Error(w, "failed to get recipe data", http.StatusInternalServerError)
		return
	case err == api_errors.ForbiddenOperation:
		http.Error(w, "forbidden operation", http.StatusForbidden)
		return
	case err.Error() == "No dimension name found":
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}

func auditActionFailure(ctx context.Context, auditedAction string, auditedResult string, err error, logData log.Data) {
	if logData == nil {
		logData = log.Data{}
	}

	logData["auditAction"] = auditedAction
	logData["auditResult"] = auditedResult

	audit.LogError(ctx, errors.WithMessage(err, auditActionErr), logData)
}

func handleAuditError(ctx context.Context, w http.ResponseWriter, auditedAction string, auditedResult string, err error, logData log.Data) {
	auditActionFailure(ctx, getJobsAction, actionAttempted, err, logData)

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(internalServerError))
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
