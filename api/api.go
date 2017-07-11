package api

import (
	"net/http"

	"encoding/json"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/schema"
	"github.com/ONSdigital/dp-import-api/utils"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// ImportAPI - A restful API used to manage importing datasets to be published
type ImportAPI struct {
	dataStore DataStore
	Router    *mux.Router
	producer  chan []byte
}

// CreateImportAPI - Create the api with all the routes configured
func CreateImportAPI(dataStore DataStore, producer chan []byte) *ImportAPI {
	router := mux.NewRouter()
	api := ImportAPI{dataStore: dataStore, Router: router, producer: producer}
	// External API for florence
	api.Router.HandleFunc("/jobs", api.addJob).Methods("POST")
	api.Router.HandleFunc("/jobs/{jobId}/file", api.addUploadedFile).Methods("PUT")
	api.Router.HandleFunc("/jobs/{jobId}/state", api.updateState).Methods("PUT")
	api.Router.HandleFunc("/instances/{instanceId}", api.getInstance).Methods("GET")
	// Internal API
	api.Router.HandleFunc("/instances/{instanceId}/events", api.addEvent).Methods("PUT")
	api.Router.HandleFunc("/instances/{instanceId}/dimensions", api.addDimension).Methods("PUT")
	api.Router.HandleFunc("/instances/{instanceId}/dimensions/{nodeName}/nodeId", api.addNodeID).Methods("PUT")
	api.Router.HandleFunc("/instances/{instanceId}/dimensions", api.getDimension).Methods("GET")

	return &api
}

func (api *ImportAPI) addJob(w http.ResponseWriter, r *http.Request) {
	newJob, error := models.CreateJob(r.Body)
	if error != nil {
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

func (api *ImportAPI) updateState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	jobState, err := models.CreateJobState(r.Body)
	if err != nil {
		log.Error(err, log.Data{"jobState": jobState, "jobID": jobID})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
	}
	dataStoreError := api.dataStore.UpdateJobState(jobID, jobState)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"jobState": jobState, "jobID": jobID})
		setErrorCode(w, dataStoreError)
		return
	}
	if jobState.State == "submitted" {
		message, error := api.dataStore.BuildPublishDatasetMessage(jobID)
		if error != nil {
			log.Error(error, log.Data{"jobState": jobState, "jobID": jobID})
			setErrorCode(w, error)
			return
		}
		bytes, avroError := schema.PublishDataset.Marshal(message)
		if avroError != nil {
			log.Error(avroError, log.Data{"recipe": message.Recipe, "byte": bytes})
			setErrorCode(w, avroError)
			return
		}
		api.producer <- bytes
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

func (api *ImportAPI) addDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	dimension, error := models.CreateDimension(r.Body)
	if error != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddDimension(instanceID, dimension)
	if dataStoreError != nil {
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

func (api *ImportAPI) addNodeID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Debug("", log.Data{"mux": vars})
	instanceID := vars["instanceId"]
	nodeName := vars["nodeName"]

	dimension, error := models.CreateDimension(r.Body)
	if error != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddNodeID(instanceID, nodeName, dimension)
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
