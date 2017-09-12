package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/dataset/interface"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/jobqueue"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

const internalError = "Internal server error"

// ImportAPI is a restful API used to manage importing datasets to be published
type ImportAPI struct {
	host       string
	dataStore  datastore.DataStore
	router     *mux.Router
	jobQueue   jobqueue.JobQueue
	datasetAPI dataset.DatasetAPIer
}

// CreateImportAPI returns the api with all the routes configured
func CreateImportAPI(host string, router *mux.Router, dataStore datastore.DataStore, jobQueue jobqueue.JobQueue, secretKey string, datasetAPI dataset.DatasetAPIer) *ImportAPI {
	api := ImportAPI{host: host, dataStore: dataStore, router: router, jobQueue: jobQueue, datasetAPI: datasetAPI}
	auth := NewAuthenticator(secretKey, "internal-token")
	// External API for florence
	api.router.Path("/jobs").Methods("POST").HandlerFunc(api.addJob)
	api.router.Path("/jobs").Methods("GET").HandlerFunc(api.getJobs).Queries()
	api.router.Path("/jobs/{job_id}").Methods("GET").HandlerFunc(api.getJob)
	api.router.Path("/jobs/{job_id}").Methods("PUT").HandlerFunc(auth.ManualCheck(api.updateJob))
	api.router.Path("/jobs/{job_id}/files").Methods("PUT").HandlerFunc(api.addUploadedFile)
	return &api
}

func (api *ImportAPI) addJob(w http.ResponseWriter, r *http.Request) {
	newJob, err := models.CreateJob(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err, log.Data{})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	if err = newJob.Validate(); err != nil {
		log.Error(err, log.Data{})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	selfURL := api.host + "/jobs/{job_id}"
	jobInstance, err := api.dataStore.AddJob(newJob, selfURL, api.datasetAPI)
	if err != nil {
		log.Error(err, log.Data{"job": newJob})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(jobInstance)
	if err != nil {
		log.Error(err, log.Data{"job": newJob})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"instance": jobInstance})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	log.Info("created new import job", log.Data{"job": jobInstance})
}

func (api *ImportAPI) getJobs(w http.ResponseWriter, r *http.Request) {
	filtersQuery := r.URL.Query().Get("job_states")
	var filterList []string
	if filtersQuery == "" {
		filterList = nil
	} else {
		filterList = strings.Split(filtersQuery, ",")
	}
	jobs, err := api.dataStore.GetJobs(filterList)
	if err != nil {
		log.Error(err, log.Data{})
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
	jobID := vars["job_id"]
	job, err := api.dataStore.GetJob(jobID)
	if err != nil {
		log.Error(err, log.Data{})
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
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	log.Info("Returning a list of import jobs", log.Data{"job_id": jobID})
}

func (api *ImportAPI) addUploadedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]
	uploadedFile, err := models.CreateUploadedFile(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err, log.Data{"file": uploadedFile, "job_id": jobID})
		http.Error(w, "bad client request received", http.StatusBadRequest)
		return
	}
	selfURL := api.host + "/jobs/" + jobID
	if _, err = api.dataStore.AddUploadedFile(jobID, uploadedFile, api.datasetAPI, selfURL); err != nil {
		log.Error(err, log.Data{"file": uploadedFile, "job_id": jobID})
		setErrorCode(w, err)
		return
	}
	log.Info("added uploaded file to job", log.Data{"job_id": jobID, "file": uploadedFile})
}

func (api *ImportAPI) updateJob(w http.ResponseWriter, r *http.Request, isAuth bool) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]
	job, err := models.CreateJob(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err, log.Data{"job": job, "job_id": jobID, "is_auth": isAuth})
		http.Error(w, "bad client request received", http.StatusBadRequest)
	}

	err = api.dataStore.UpdateJob(jobID, job, isAuth)
	if err != nil {
		log.Error(err, log.Data{"job": job, "job_id": jobID})
		setErrorCode(w, err)
		return
	}
	log.Info("job updated", log.Data{"job": job, "job_id": jobID, "is_auth": isAuth})
	if job.State == "submitted" {
		tasks, err := api.dataStore.PrepareJob(api.datasetAPI, jobID)
		if err != nil {
			log.Error(err, log.Data{"jobState": job, "job_id": jobID, "isAuth": isAuth})
			setErrorCode(w, err)
			return
		}
		err = api.jobQueue.Queue(tasks)
		if err != nil {
			log.Error(err, log.Data{"tasks": tasks})
			setErrorCode(w, err)
			return
		}
		log.Info("import job was queued", log.Data{"job": job, "job_id": jobID, "is_auth": isAuth})
	}
}

func setErrorCode(w http.ResponseWriter, err error) {
	switch {
	case err == api_errors.JobNotFoundError:
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	case err == api_errors.ForbiddenOperation:
		http.Error(w, "Forbidden operation", http.StatusForbidden)
		return
	case err.Error() == "No dimension name found":
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
