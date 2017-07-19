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
	api.Router.HandleFunc("/instances/{instanceId}/dimensions", api.getDimension).Methods("GET")

	return &api
}

func (api *ImportAPI) addJob(w http.ResponseWriter, r *http.Request) {
	newJob, error := models.CreateJob(r.Body)
	if error != nil || newJob.Validate() != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	jobInstance, dataStoreError := api.dataStore.AddJob(newJob)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"newJob": newJob})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	bytes, error := json.Marshal(jobInstance)
	if error != nil {
		log.Error(error, log.Data{"Job": newJob})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, writeError := w.Write(bytes)
	if writeError != nil {
		log.Error(writeError, log.Data{"jobInstance": jobInstance})
	}
}

func (api *ImportAPI) addUploadedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	uploadedFile, error := models.CreateUploadedFile(r.Body)
	if error != nil {
		log.Error(error, log.Data{"uploadedFile": uploadedFile, "jobID": jobID})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddUploadedFile(jobID, uploadedFile)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"uploadedFile": uploadedFile, "instanceId": jobID})
		setErrorCode(w, dataStoreError)
		return
	}
}

func (api *ImportAPI) updateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	job, err := models.CreateJob(r.Body)
	if err != nil {
		log.Error(err, log.Data{"jobState": job, "jobID": jobID})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
	}
	dataStoreError := api.dataStore.UpdateJobState(jobID, job)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"jobState": job, "jobID": jobID})
		setErrorCode(w, dataStoreError)
		return
	}
	if job.State == "submitted" {
		task, error := api.dataStore.BuildPublishDatasetMessage(jobID)
		if error != nil {
			log.Error(error, log.Data{"jobState": job, "jobID": jobID})
			setErrorCode(w, error)
			return
		}
		queueError := api.jobQueue.Queue(task)
		if queueError != nil {
			log.Error(queueError, log.Data{"task": task})
			setErrorCode(w, queueError)
		}
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, writeError := w.Write(bytes)
	if writeError != nil {
		log.Error(writeError, log.Data{"instanceState": instanceState})
	}
}

func (api *ImportAPI) addEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	event, error := models.CreateEvent(r.Body)
	if error != nil {
		log.Error(error, log.Data{"event": event})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddEvent(instanceID, event)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"event": event})
		setErrorCode(w, dataStoreError)
		return
	}
}

func (api *ImportAPI) getDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimensions, dataStoreError := api.dataStore.GetDimension(instanceID)
	if dataStoreError != nil {
		setErrorCode(w, dataStoreError)
		return
	}
	bytes, error := json.Marshal(dimensions)
	if error != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, writeError := w.Write(bytes)
	if writeError != nil {
		log.Error(writeError, log.Data{"dimensions": dimensions})
	}
}

func (api *ImportAPI) addDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimensionName := vars["dimension_name"]
	dimensionValue := vars["value"]
	if dimensionName == "" || dimensionValue == "" {
		log.Error(fmt.Errorf("Missing parameters"), log.Data{"name": dimensionName, "value": dimensionValue})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dimension := models.Dimension{Name: dimensionName, Value: dimensionValue}
	dataStoreError := api.dataStore.AddDimension(instanceID, &dimension)
	if dataStoreError != nil {
		setErrorCode(w, dataStoreError)
		return
	}
}

func (api *ImportAPI) addNodeID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimensionName := vars["dimension_name"]
	nodeValue := vars["value"]
	if dimensionName == "" || nodeValue == "" {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dimension := models.Dimension{NodeID: nodeValue}
	dataStoreError := api.dataStore.AddNodeID(instanceID, dimensionName, &dimension)
	setErrorCode(w, dataStoreError)

}

func setErrorCode(w http.ResponseWriter, err error) {
	switch {
	case err == utils.JobNotFoundError:
		http.Error(w, "Import job not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
