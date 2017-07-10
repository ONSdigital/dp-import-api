package api

import (
	"net/http"

	"encoding/json"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/utils"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	//"github.com/ONSdigital/go-ns/avro"
	"github.com/ONSdigital/dp-import-api/schema"
)
// ImportAPI - .....
type ImportAPI struct {
	dataStore DataStore
	Router    *mux.Router
	Output    chan []byte
}

// CreateImportAPI - ....
func CreateImportAPI(dataStore DataStore) *ImportAPI {
	router := mux.NewRouter()
	api := ImportAPI{dataStore: dataStore, Router: router}
	// External API for florence
	api.Router.HandleFunc("/job", api.addJob).Methods("POST")
	api.Router.HandleFunc("/job/{jobId}/s3file", api.addUploadedFile).Methods("PUT")
	api.Router.HandleFunc("/job/{jobId}/state", api.updateState).Methods("PUT")
	api.Router.HandleFunc("/import/{instanceId}", api.getInstance).Methods("GET")
	// Internal API
	api.Router.HandleFunc("/import/{instanceId}/events", api.addEvent).Methods("PUT")
	api.Router.HandleFunc("/import/{instanceId}/dimensions", api.addDimension).Methods("PUT")
	api.Router.HandleFunc("/import/{instanceId}/dimensions/{nodeName}/nodeId", api.addNodeID).Methods("PUT")
	api.Router.HandleFunc("/import/{instanceId}/dimensions", api.getDimension).Methods("GET")

	return &api
}

func (api *ImportAPI) addJob(w http.ResponseWriter, r *http.Request) {
	message, error := models.CreateJob(r.Body)
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
	setJSONContentType(w)
	 _, writeError := w.Write(bytes)
	if writeError != nil {
		log.Error(writeError, log.Data{"jobInstance": jobInstance})
	}
}

func (api *ImportAPI) addUploadedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobID"]
	message, error := models.CreateUploadedFile(r.Body)
	if error != nil {
		log.Error(error, log.Data{"message": message, "jobID": jobID})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddUploadedFile(jobID, message)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"message": message, "instanceId": jobID})
		setErrorCode(w, dataStoreError)
		return
	}
}

func (api *ImportAPI) updateState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobID"]
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
		byte, avroError := schema.PublishDataset.Marshal(&message)
		if avroError != nil {
			log.Error(avroError, log.Data{"message": message, "byte": byte})
			setErrorCode(w, avroError)
			return
		}
	}
}

func (api *ImportAPI) getInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceID"]
	importJob, dataStoreError := api.dataStore.GetInstance(instanceID)
	if dataStoreError != nil {
		log.Error(dataStoreError, log.Data{"instanceID": instanceID})
		setErrorCode(w, dataStoreError)
		return
	}

	bytes, error := json.Marshal(importJob)
	if error != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, writeError := w.Write(bytes)
	if writeError != nil {
		log.Error(writeError, log.Data{"importJob": importJob})
	}
}

func (api *ImportAPI) addEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceID"]
	message, error := models.CreateEvent(r.Body)
	if error != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddEvent(instanceID, message)
	if dataStoreError != nil {
		setErrorCode(w, dataStoreError)
		return
	}
}

func (api *ImportAPI) addDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceID"]
	message, error := models.CreateDimension(r.Body)
	if error != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddDimension(instanceID, message)
	if dataStoreError != nil {
		setErrorCode(w, dataStoreError)
		return
	}
}

func (api *ImportAPI) getDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceID"]
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

	message, error := models.CreateDimension(r.Body)
	if error != nil {
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dataStoreError := api.dataStore.AddNodeID(instanceID, nodeName, message)
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
