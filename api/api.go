package api

import (
	"net/http"

	"encoding/json"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

type ImportAPI struct {
	dataStore datastore.DataStore
	Router    *mux.Router
}

func CreateImportAPI(dataStore datastore.DataStore) *ImportAPI {
	router := mux.NewRouter()
	api := ImportAPI{dataStore: dataStore, Router: router}
	// External API for florence
	api.Router.HandleFunc("/job", api.addJob).Methods("POST")
	api.Router.HandleFunc("/job/{jobId}/s3file", api.addS3File).Methods("PUT")
	api.Router.HandleFunc("/job/{jobId}/state", api.updateState).Methods("PUT")
	api.Router.HandleFunc("/import/{instanceId}", api.getInstance).Methods("GET")
	// Internal API
	api.Router.HandleFunc("/import/{instanceId}/events", api.addEvent).Methods("PUT")
	api.Router.HandleFunc("/import/{instanceId}/dimensions", api.addDimension).Methods("PUT")
	api.Router.HandleFunc("/import/{instanceId}/dimensions/{nodeName}/nodeId", api.addNodeId).Methods("PUT")
	api.Router.HandleFunc("/import/{instanceId}/dimensions", api.getDimension).Methods("GET")

	return &api
}

func (api *ImportAPI) addJob(w http.ResponseWriter, r *http.Request) {
	message, error := models.CreateImportJob(r.Body)
	if error != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	jobInstance, dataStoreError := api.dataStore.AddJob(message)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"message": message})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	bytes, error := json.Marshal(jobInstance)
	if error != nil {
		log.Error(error, log.Data{"message": message})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	setJsonContentType(w)
	w.Write(bytes)
}

func (api *ImportAPI) addS3File(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobId := vars["jobId"]
	message, error := models.CreateS3File(r.Body)
	if error != nil {
		log.Error(error, log.Data{"message": message, "jobId": jobId})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddS3File(jobId, message)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"message": message, "instanceId": jobId})
		setErrorCode(w, dataStoreError)
		return
	}
}

func (api *ImportAPI) updateState(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//jobId := vars["jobId"]

}

func (api *ImportAPI) getInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	importJob, dataStoreError := api.dataStore.GetInstance(instanceId)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"instanceId": instanceId})
		setErrorCode(w, dataStoreError)
		return
	}

	bytes, error := json.Marshal(importJob)
	if error != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	setJsonContentType(w)
	w.Write(bytes)
}

func (api *ImportAPI) addEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	message, error := models.CreateEvent(r.Body)
	if error != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddEvent(instanceId, message)
	if dataStoreError != nil {
		setErrorCode(w, dataStoreError)
		return
	}
}

func (api *ImportAPI) addDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	message, error := models.CreateDimension(r.Body)
	if error != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddDimension(instanceId, message)
	if dataStoreError != nil {
		setErrorCode(w, dataStoreError)
		return
	}
}

func (api *ImportAPI) getDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	dimensions, dataStoreError := api.dataStore.GetDimension(instanceId)
	if dataStoreError != nil {
		setErrorCode(w, dataStoreError)
		return
	}
	bytes, error := json.Marshal(dimensions)
	if error != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	setJsonContentType(w)
	w.Write(bytes)
}

func (api *ImportAPI) addNodeId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Debug("", log.Data{"mux": vars})
	instanceId := vars["instanceId"]
	nodeName := vars["nodeName"]

	message, error := models.CreateDimension(r.Body)
	if error != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddNodeId(instanceId, nodeName, message)
	setErrorCode(w, dataStoreError)

}

func setErrorCode(w http.ResponseWriter, err error) {
	switch {
	case err == datastore.JobNotFoundError:
		http.Error(w, "Import job not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func setJsonContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
