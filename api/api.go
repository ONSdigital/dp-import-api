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
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
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

func (api *ImportAPI) getJobs(w http.ResponseWriter, r *http.Request) {
	if err := api.auditor.Record(r.Context(), getJobsAction, actionAttempted, nil); err != nil {
		handleAuditError(w, err, nil)
		return
	}

	filtersQuery := r.URL.Query().Get("state")
	var filterList []string
	if filtersQuery == "" {
		filterList = nil
	} else {
		filterList = strings.Split(filtersQuery, ",")
	}

	jobs, err := api.dataStore.GetJobs(filterList)
	if err != nil {
		log.Error(err, log.Data{})
		if err := api.auditor.Record(r.Context(), getJobsAction, actionUnsuccessful, nil); err != nil {
			handleAuditError(w, err, nil)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(jobs)
	if err != nil {
		log.Error(err, log.Data{"Jobs": jobs})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)

	if err := api.auditor.Record(r.Context(), getJobsAction, actionSuccessful, nil); err != nil {
		handleAuditError(w, err, nil)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	log.Info("Returning a list of import jobs", log.Data{"filter": filtersQuery})
}

func (api *ImportAPI) getJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}
	auditParams := common.Params{jobIDKey: jobID}

	if err := api.auditor.Record(r.Context(), getJobAction, actionAttempted, auditParams); err != nil {
		handleAuditError(w, err, logData)
		return
	}

	job, err := api.dataStore.GetJob(jobID)
	if err != nil {
		if err := api.auditor.Record(r.Context(), getJobAction, actionUnsuccessful, auditParams); err != nil {
			handleAuditError(w, err, logData)
			return
		}
		log.Error(err, logData)
		setErrorCode(w, err)
		return
	}
	bytes, err := json.Marshal(job)
	if err != nil {
		log.Error(err, log.Data{"job": job})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	if err := api.auditor.Record(r.Context(), getJobAction, actionSuccessful, auditParams); err != nil {
		handleAuditError(w, err, logData)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	log.Info("Returning an import job", log.Data{jobIDKey: jobID})
}

func (api *ImportAPI) addJob(w http.ResponseWriter, r *http.Request) {
	job, err := models.CreateJob(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err, log.Data{})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}

	logData := log.Data{"recipeID": job.RecipeID}
	auditParams := common.Params{"recipeID": job.RecipeID}
	if err := api.auditor.Record(r.Context(), addJobAction, actionAttempted, auditParams); err != nil {
		handleAuditError(w, err, logData)
		return
	}

	createdJob, err := api.jobService.CreateJob(r.Context(), job)
	if err != nil {
		if err := api.auditor.Record(r.Context(), addJobAction, actionUnsuccessful, auditParams); err != nil {
			handleAuditError(w, err, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	logData["job"] = job
	bytes, err := json.Marshal(createdJob)
	if err != nil {
		log.Error(err, logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	auditParams["createdJobID"] = createdJob.ID
	if err := api.auditor.Record(r.Context(), addJobAction, actionSuccessful, auditParams); err != nil {
		handleAuditError(w, err, logData)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(bytes); err != nil {
		log.Error(err, log.Data{"instance": createdJob})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	log.Info("created new import job", log.Data{"job": createdJob})
}

func (api *ImportAPI) addUploadedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}
	auditParams := common.Params{jobIDKey: jobID}

	if err := api.auditor.Record(r.Context(), uploadFileAction, actionAttempted, auditParams); err != nil {
		handleAuditError(w, err, logData)
		return
	}

	uploadedFile, err := models.CreateUploadedFile(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err, logData)
		http.Error(w, "bad client request received", http.StatusBadRequest)
		return
	}

	logData["file"] = uploadedFile
	auditParams["fileAlias"] = uploadedFile.AliasName
	auditParams["fileURL"] = uploadedFile.URL

	if err := api.dataStore.AddUploadedFile(jobID, uploadedFile); err != nil {
		if err := api.auditor.Record(r.Context(), uploadFileAction, actionUnsuccessful, auditParams); err != nil {
			handleAuditError(w, err, logData)
			return
		}
		log.ErrorC("error while adding up", err, logData)
		setErrorCode(w, err)
		return
	}
	if err := api.auditor.Record(r.Context(), uploadFileAction, actionSuccessful, auditParams); err != nil {
		handleAuditError(w, err, logData)
		return
	}
	log.Info("added uploaded file to job", log.Data{jobIDKey: jobID, "file": uploadedFile})
}

func (api *ImportAPI) updateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}
	auditParams := common.Params{jobIDKey: jobID}

	if err := api.auditor.Record(r.Context(), updateJobAction, actionAttempted, auditParams); err != nil {
		handleAuditError(w, err, logData)
		return
	}

	job, err := models.CreateJob(r.Body)
	defer r.Body.Close()
	if err != nil {
		logData["job"] = job
		log.Error(err, logData)
		http.Error(w, "bad client request received", http.StatusBadRequest)
		return
	}

	if err := api.jobService.UpdateJob(r.Context(), jobID, job); err != nil {
		log.ErrorC("error updating job", err, logData)
		if err := api.auditor.Record(r.Context(), updateJobAction, actionUnsuccessful, auditParams); err != nil {
			handleAuditError(w, err, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	if err := api.auditor.Record(r.Context(), updateJobAction, actionSuccessful, auditParams); err != nil {
		handleAuditError(w, err, logData)
		return
	}
	log.Info("job update completed successfully", logData)
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
	case err == job.ErrCreateInstanceFailed:
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

func handleAuditError(w http.ResponseWriter, err error, logData log.Data) {
	log.ErrorC(auditError, err, logData)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(internalServerError))
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
