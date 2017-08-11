package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

const internalError = "Internal server error"

// ImportAPI is a restful API used to manage importing datasets to be published
type ImportAPI struct {
	host      string
	dataStore DataStore
	router    *mux.Router
	jobQueue  JobQueue
}

// CreateImportAPI returns the api with all the routes configured
func CreateImportAPI(host string, router *mux.Router, dataStore DataStore, jobQueue JobQueue, secretKey string) *ImportAPI {
	api := ImportAPI{host: host, dataStore: dataStore, router: router, jobQueue: jobQueue}
	auth := NewAuthenticator(secretKey, "internal-token")
	// External API for florence
	api.router.Path("/jobs").Methods("POST").HandlerFunc(api.addJob)
	api.router.Path("/jobs").Methods("GET").HandlerFunc(api.getJobs).Queries()
	api.router.Path("/jobs/{job_id}").Methods("GET").HandlerFunc(api.getJob)
	api.router.Path("/jobs/{job_id}").Methods("PUT").HandlerFunc(auth.ManualCheck(api.updateJob))
	api.router.Path("/jobs/{job_id}/files").Methods("PUT").HandlerFunc(api.addUploadedFile)
	api.router.Path("/instances").Methods("GET").HandlerFunc(api.getInstances)
	api.router.Path("/instances/{instance_id}").Methods("GET").HandlerFunc(api.getInstance)
	api.router.Path("/instances/{instance_id}/dimensions").Methods("GET").HandlerFunc(api.getDimensions)
	api.router.Path("/instances/{instance_id}/dimensions/{dimension_name}/options").Methods("GET").HandlerFunc(api.getDimensionValues)
	// Internal API
	api.router.Path("/instances/{instance_id}").Methods("PUT").HandlerFunc(auth.Check(api.updateInstance))
	api.router.Path("/instances/{instance_id}/events").Methods("PUT").HandlerFunc(auth.Check(api.addEvent))
	api.router.Path("/instances/{instance_id}/dimensions/{dimension_name}/options/{value}").
		Methods("PUT").HandlerFunc(auth.Check(api.addDimension))
	api.router.Path("/instances/{instance_id}/dimensions/{dimension_name}/options/{dimension_value}/node_id/{node_id}").Methods("PUT").
		HandlerFunc(auth.Check(api.addNodeID))
	api.router.Path("/instances/{instance_id}/inserted_observations/{inserted_observations}").Methods("PUT").HandlerFunc(auth.Check(api.addInsertedObservations))

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
	jobs, err := api.dataStore.GetJobs(api.host, filterList)
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
	job, err := api.dataStore.GetJob(api.host, jobID)
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
	err = api.dataStore.AddUploadedFile(jobID, uploadedFile)
	if err != nil {
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

	err = api.dataStore.UpdateJobState(jobID, job, isAuth)
	if err != nil {
		log.Error(err, log.Data{"job": job, "job_id": jobID})
		setErrorCode(w, err)
		return
	}
	log.Info("job updated", log.Data{"job": job, "job_id": jobID, "is_auth": isAuth})
	if job.State == "submitted" {
		task, err := api.dataStore.PrepareImportJob(jobID)
		if err != nil {
			log.Error(err, log.Data{"jobState": job, "job_id": jobID, "isAuth": isAuth})
			setErrorCode(w, err)
			return
		}
		err = api.jobQueue.Queue(task)
		if err != nil {
			log.Error(err, log.Data{"task": task})
			setErrorCode(w, err)
			return
		}
		log.Info("import job was queued", log.Data{"job": job, "job_id": jobID, "is_auth": isAuth})
	}
}

func (api *ImportAPI) getInstances(w http.ResponseWriter, r *http.Request) {
	filtersQuery := r.URL.Query().Get("instance_states")
	var filterList []string
	if filtersQuery != "" {
		filterList = strings.Split(filtersQuery, ",")
	}
	instances, err := api.dataStore.GetInstances(api.host, filterList)
	if err != nil {
		log.Error(err, log.Data{})
		setErrorCode(w, err)
		return
	}
	bytes, err := json.Marshal(instances)
	if err != nil {
		log.Error(err, log.Data{})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{})
		log.Error(err, log.Data{"instance": instances})
		return
	}
	log.Info("returning instances information", log.Data{})
}

func (api *ImportAPI) getInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instance_id"]
	instanceState, err := api.dataStore.GetInstance(api.host, instanceID)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID})
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
		log.Error(err, log.Data{"instance_state": instanceState})
		return
	}
	log.Info("returning instance information", log.Data{"instance_id": instanceID})
}

func (api *ImportAPI) updateInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instance_id"]
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
	instanceID := vars["instance_id"]
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
	instanceID := vars["instance_id"]
	dimensions, err := api.dataStore.GetDimensions(instanceID)
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
	instanceID := vars["instance_id"]
	dimensionName := vars["dimension_name"]
	dimensionValue := vars["value"]
	if dimensionName == "" || dimensionValue == "" {
		log.Error(errors.New("missing parameters"), log.Data{"instance_id": instanceID, "name": dimensionName, "value": dimensionValue})
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

func (api *ImportAPI) getDimensionValues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instance_id"]
	dimensionName := vars["dimension_name"]
	dimensionValues, err := api.dataStore.GetDimensionValues(instanceID, dimensionName)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "name": dimensionName})
		setErrorCode(w, err)
		return
	}
	bytes, err := json.Marshal(dimensionValues)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "dimension_values": dimensionValues})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	setJSONContentType(w)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "dimension_values": dimensionValues})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	log.Info("returned dimension values", log.Data{"instance_id": instanceID})
}

func (api *ImportAPI) addNodeID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instance_id"]
	dimensionValue := vars["dimension_value"]
	dimensionName := vars["dimension_name"]
	nodeValue := vars["node_id"]
	if dimensionName == "" || nodeValue == "" || dimensionValue == "" {
		log.Error(errors.New("missing parameters"), log.Data{"instance_id": instanceID, "dimension_name": dimensionName, "dimension_value": nodeValue})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	dimension := models.Dimension{Name: dimensionName, Value: dimensionValue, NodeID: nodeValue}
	err := api.dataStore.AddNodeID(instanceID, &dimension)
	if err != nil {
		log.Error(err, log.Data{"instance_id": instanceID, "dimension": dimension})
		setErrorCode(w, err)
		return
	}
	log.Info("added node id", log.Data{"instance_id": instanceID, "dimension": dimension})
}

func (api *ImportAPI) addInsertedObservations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instance_id"]
	insertedObservations, err := strconv.Atoi(vars["inserted_observations"])
	if insertedObservations == 0 || err != nil {
		log.Error(errors.New("missing / invalid parameters"), log.Data{"instance_id": instanceID, "inserted_observations": insertedObservations})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}
	err = api.dataStore.UpdateObservationCount(instanceID, insertedObservations)
	if err != nil {
		setErrorCode(w, err)
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
