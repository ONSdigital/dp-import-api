package api

import (
	"net/http"

	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/utils"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

var internalError = "Internal server error"

// ImportAPI - A restful API used to manage importing datasets to be published
type ImportAPI struct {
	dataStore DataStore
	Router    *mux.Router
	jobQueue  JobQueue
}

// CreateImportAPI - Create the api with all the routes configured
func CreateImportAPI(dataStore DataStore, jobQueue JobQueue) *ImportAPI {
	router := mux.NewRouter()
	api := ImportAPI{dataStore: dataStore, Router: router, jobQueue: jobQueue}
	// External API for florence
	api.Router.HandleFunc("/jobs", api.addJob).Methods("POST")
	api.Router.HandleFunc("/jobs/{jobId}", api.updateJob).Methods("PUT")
	api.Router.HandleFunc("/jobs/{jobId}/files", api.addUploadedFile).Methods("PUT")
	api.Router.HandleFunc("/instances/{instanceId}", api.getInstance).Methods("GET")
	// Internal API
	api.Router.HandleFunc("/instances/{instanceId}/events", api.addEvent).Methods("PUT")
	api.Router.HandleFunc("/instances/{instanceId}/dimensions/{dimension_name}/options/{value}", api.addDimension).Methods("PUT")
	api.Router.HandleFunc("/instances/{instanceId}/dimensions/{dimension_name}/nodeid/{value}", api.addNodeID).Methods("PUT")
	api.Router.HandleFunc("/instances/{instanceId}/dimensions", api.getDimensions).Methods("GET")

	return &api
}

func (api *ImportAPI) addJob(w http.ResponseWriter, r *http.Request) {
	newJob, error := models.CreateJob(r.Body)
	if error != nil  {
		log.Error(error,log.Data{})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	if validationError := newJob.Validate();  validationError != nil  {
		log.Error(validationError,log.Data{})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	jobInstance, dataStoreError := api.dataStore.AddJob(newJob)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"newJob": newJob})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	bytes, error := json.Marshal(jobInstance)
	if error != nil {
		log.Error(error, log.Data{"Job": newJob})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, writeError := w.Write(bytes)
	if writeError != nil {
		log.Error(writeError, log.Data{"jobInstance": jobInstance})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	log.Info("created new import job", log.Data{"job": jobInstance})
	w.WriteHeader(http.StatusCreated)
}

func (api *ImportAPI) addUploadedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	uploadedFile, error := models.CreateUploadedFile(r.Body)
	if error != nil {
		log.Error(error, log.Data{"file": uploadedFile, "job_id": jobID})
		http.Error(w, "bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddUploadedFile(jobID, uploadedFile)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"file": uploadedFile, "job_id": jobID})
		setErrorCode(w, dataStoreError)
		return
	}
	log.Info("added uploaded file to job", log.Data{"job_id":jobID, "file":uploadedFile})
}

func (api *ImportAPI) updateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	job, err := models.CreateJob(r.Body)
	if err != nil {
		log.Error(err, log.Data{"jobState": job, "job_id": jobID})
		http.Error(w, "bad client request received", http.StatusBadRequest)
	}
	dataStoreError := api.dataStore.UpdateJobState(jobID, job)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"job": job, "job_id": jobID})
		setErrorCode(w, dataStoreError)
		return
	}
	log.Info("job updated", log.Data{"job_id":jobID, "updates": job})
	if job.State == "submitted" {
		task, error := api.dataStore.BuildPublishDatasetMessage(jobID)
		if error != nil {
			log.Error(error, log.Data{"job": job, "job_id": jobID})
			setErrorCode(w, error)
			return
		}
		queueError := api.jobQueue.Queue(task)
		if queueError != nil {
			log.Error(queueError, log.Data{"task": task})
			setErrorCode(w, queueError)
			return
		}
		log.Info("import job was queued", log.Data{"job_id":jobID})
	}
}

func (api *ImportAPI) getInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	instanceState, dataStoreError := api.dataStore.GetInstance(instanceID)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"instanceID": instanceID})
		setErrorCode(w, dataStoreError)
		return
	}
	bytes, error := json.Marshal(instanceState)
	if error != nil {
		log.Error(error, log.Data{"instance_id":instanceID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, writeError := w.Write(bytes)
	if writeError != nil {
		log.Error(error, log.Data{"instance_id":instanceID})
		log.Error(writeError, log.Data{"instanceState": instanceState})
		return
	}
	log.Info("returning instance information", log.Data{"instance_id":instanceID})
}

func (api *ImportAPI) addEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	event, error := models.CreateEvent(r.Body)
	if error != nil {
		log.Error(error, log.Data{"instance_id":instanceID, "event":event})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddEvent(instanceID, event)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"instance_id":instanceID, "event":event})
		setErrorCode(w, dataStoreError)
		return
	}
	log.Info("event added to instance", log.Data{"instance_id":instanceID, "event":event})
	w.WriteHeader(http.StatusCreated)
}

func (api *ImportAPI) getDimensions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimensions, dataStoreError := api.dataStore.GetDimension(instanceID)
	if dataStoreError != nil {
		setErrorCode(w, dataStoreError)
		return
	}
	bytes, error := json.Marshal(dimensions)
	if error != nil {
		log.Error(error, log.Data{"instance_id":instanceID, "dimensions": dimensions})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, writeError := w.Write(bytes)
	if writeError != nil {
		log.Error(error, log.Data{"instance_id":instanceID, "dimensions": dimensions})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	log.Info("returned dimensions", log.Data{"instance_id":instanceID})
}

func (api *ImportAPI) addDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimensionName := vars["dimension_name"]
	dimensionValue := vars["value"]
	if dimensionName == "" || dimensionValue == "" {
		log.Error(fmt.Errorf("Missing parameters"), log.Data{"instance_id":instanceID, "name": dimensionName, "value": dimensionValue})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dimension := models.Dimension{Name: dimensionName, Value: dimensionValue}
	dataStoreError := api.dataStore.AddDimension(instanceID, &dimension)
	if dataStoreError != nil {
		setErrorCode(w, dataStoreError)
		return
	}
	log.Info("dimension added", log.Data{"instance_id":instanceID, "name": dimensionName, "value": dimensionValue})
}

func (api *ImportAPI) addNodeID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimensionName := vars["dimension_name"]
	nodeValue := vars["value"]
	if dimensionName == "" || nodeValue == "" {
		log.Error(fmt.Errorf("Missing parameters"), log.Data{"instance_id":instanceID, "name": dimensionName, "nodeValue": nodeValue})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dimension := models.Dimension{NodeID: nodeValue}
	dataStoreError := api.dataStore.AddNodeID(instanceID, dimensionName, &dimension)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"instance_id":instanceID, "name": dimensionName, "nodeValue": nodeValue})
		setErrorCode(w, dataStoreError)
		return
	}
    log.Info("added node id", log.Data{"instance_id":instanceID, "name": dimensionName, "nodeValue": nodeValue})
}

func setErrorCode(w http.ResponseWriter, err error) {
	switch {
	case err == utils.JobNotFoundError:
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
