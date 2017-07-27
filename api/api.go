package api

import (
	"net/http"

	"encoding/json"
	"errors"
	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

const internalError = "Internal server error"

// ImportAPI - A restful API used to manage importing datasets to be published
type ImportAPI struct {
	host      string
	dataStore DataStore
	router    *mux.Router
	jobQueue  JobQueue
}

// CreateImportAPI - Createerr the api with all the routes configured
func CreateImportAPI(host string, router *mux.Router ,dataStore DataStore, jobQueue JobQueue) *ImportAPI {
	api := ImportAPI{host: host, dataStore: dataStore, router: router, jobQueue: jobQueue}
	// External API for florence
	api.router.Path("/jobs").Methods("POST").HandlerFunc(api.addJob)
	api.router.Path("/jobs/{jobId}").Methods("PUT").HandlerFunc(api.updateJob)
	api.router.Path("/jobs/{jobId}/files").Methods("PUT").HandlerFunc( api.addUploadedFile)
	api.router.Path("/instances/{instanceId}", ).Methods("GET").HandlerFunc(api.getInstance)
	// Internal API
	api.router.Path("/instances/{instanceId}").Methods("PUT").HandlerFunc(api.updateInstance)
	api.router.Path("/instances/{instanceId}/events").Methods("PUT").HandlerFunc(api.addEvent)
	api.router.Path("/instances/{instanceId}/dimensions/{dimension_name}/options/{value}").Methods("PUT").HandlerFunc(api.addDimension)
	api.router.Path("/instances/{instanceId}/dimensions/{dimension_name}/nodeid/{value}").Methods("PUT").HandlerFunc(api.addNodeID)
	api.router.Path("/instances/{instanceId}/dimensions").Methods("GET").HandlerFunc(api.getDimensions)

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
	jobInstance, err := api.dataStore.AddJob(api.host, newJob)
	if err != nil {
		log.Error(err, log.Data{"newJob": newJob})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(jobInstance)
	if err != nil {
		log.Error(err, log.Data{"Job": newJob})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"jobInstance": jobInstance})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	log.Info("created new import job", log.Data{"job": jobInstance})
}

func (api *ImportAPI) addUploadedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	uploadedFile, err := models.CreateUploadedFile(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err, log.Data{"file": uploadedFile, "job_id": jobID})
		http.Error(w, "bad client request received", http.StatusBadRequest)
		return
	}
	err = api.dataStore.AddUploadedFile(jobID, uploadedFile)
	if err != nil {
		log.Error(err, log.Data{"file": uploadedFile, "job_id": jobID})
		setErrorCode(w, err)
		return
	}
	log.Info("added uploaded file to job", log.Data{"job_id": jobID, "file": uploadedFile})
}

func (api *ImportAPI) updateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	job, err := models.CreateJob(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err, log.Data{"jobState": job, "job_id": jobID})
		http.Error(w, "bad client request received", http.StatusBadRequest)
	}
	err = api.dataStore.UpdateJobState(jobID, job)
	if err != nil {
		log.Error(err, log.Data{"job": job, "job_id": jobID})
		setErrorCode(w, err)
		return
	}
	log.Info("job updated", log.Data{"job_id": jobID, "updates": job})
	if job.State == "submitted" {
		task, err := api.dataStore.BuildImportDataMessage(jobID)
		if err != nil {
			log.Error(err, log.Data{"job": job, "job_id": jobID})
			setErrorCode(w, err)
			return
		}
		err = api.jobQueue.Queue(task)
		if err != nil {
			log.Error(err, log.Data{"task": task})
			setErrorCode(w, err)
			return
		}
		log.Info("import job was queued", log.Data{"job_id": jobID})
	}
}

func (api *ImportAPI) getInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	instanceState, err := api.dataStore.GetInstance(instanceID)
	if err != nil {
		log.Error(err, log.Data{"instanceID": instanceID})
		setErrorCode(w, err)
		return
	}
	bytes, err := json.Marshal(instanceState)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID})
		log.Error(err, log.Data{"instanceState": instanceState})
		return
	}
	log.Info("returning instance information", log.Data{"instance_id": instanceID})
}

func (api *ImportAPI) updateInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	instance, err := models.CreateInstance(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "instance": instance})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	err = api.dataStore.UpdateInstance(instanceID, instance)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "instance": instance})
		setErrorCode(w, err)
		return
	}
	log.Info("instance was updated", log.Data{"instance_id": instanceID, "instance": instance})
}

func (api *ImportAPI) addEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	event, err := models.CreateEvent(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "event": event})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	err = api.dataStore.AddEvent(instanceID, event)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "event": event})
		setErrorCode(w, err)
		return
	}
	log.Info("event added to instance", log.Data{"instance_id": instanceID, "event": event})
	w.WriteHeader(http.StatusCreated)
}

func (api *ImportAPI) getDimensions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimensions, err := api.dataStore.GetDimension(instanceID)
	if err != nil {
		setErrorCode(w, err)
		return
	}
	bytes, err := json.Marshal(dimensions)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "dimensions": dimensions})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "dimensions": dimensions})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	log.Info("returned dimensions", log.Data{"instance_id": instanceID})
}

func (api *ImportAPI) addDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimensionName := vars["dimension_name"]
	dimensionValue := vars["value"]
	if dimensionName == "" || dimensionValue == "" {
		log.Error(errors.New("Missing parameters"), log.Data{"instance_id": instanceID, "name": dimensionName, "value": dimensionValue})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dimension := models.Dimension{Name: dimensionName, Value: dimensionValue}
	err := api.dataStore.AddDimension(instanceID, &dimension)
	if err != nil {
		setErrorCode(w, err)
		return
	}
	log.Info("dimension added", log.Data{"instance_id": instanceID, "name": dimensionName, "value": dimensionValue})
}

func (api *ImportAPI) addNodeID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimensionName := vars["dimension_name"]
	nodeValue := vars["value"]
	if dimensionName == "" || nodeValue == "" {
		log.Error(errors.New("Missing parameters"), log.Data{"instance_id": instanceID, "name": dimensionName, "nodeValue": nodeValue})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dimension := models.Dimension{NodeID: nodeValue}
	err := api.dataStore.AddNodeID(instanceID, dimensionName, &dimension)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "name": dimensionName, "nodeValue": nodeValue})
		setErrorCode(w, err)
		return
	}
	log.Info("added node id", log.Data{"instance_id": instanceID, "name": dimensionName, "nodeValue": nodeValue})
}

func setErrorCode(w http.ResponseWriter, err error) {
	switch {
	case err == api_errors.JobNotFoundError:
		http.Error(w, "Import job not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
