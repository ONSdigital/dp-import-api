package dataset

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rhttp"
)

// DatasetAPI aggregates a client and URL and other common data for accessing the API
type DatasetAPI struct {
	Client    *rhttp.Client
	url       string
	AuthToken string
}

// NewDatasetAPI creates an DatasetAPI object
func NewDatasetAPI(client *rhttp.Client, datasetAPIURL, datasetAPIAuthToken string) *DatasetAPI {
	return &DatasetAPI{
		Client:    client,
		url:       datasetAPIURL,
		AuthToken: datasetAPIAuthToken,
	}
}

// CreateInstance tells the Dataset API to create a Dataset instance
func (api *DatasetAPI) CreateInstance(job *models.Job, recipeInst *models.RecipeInstance) (instance *models.Instance, err error) {
	path := api.url + "/instances"
	datasetPath := api.url + "/datasets/" + recipeInst.DatasetID
	logData := log.Data{"URL": path, "job_id": job.ID, "job_url": job.Links.Self.HRef}

	var jsonUpload []byte
	if jsonUpload, err = json.Marshal(models.CreateInstance(job, recipeInst.DatasetID, datasetPath, recipeInst.CodeLists)); err != nil {
		log.ErrorC("CreateInstance marshal", err, logData)
		return
	}
	logData["jsonUpload"] = jsonUpload
	jsonResult, httpCode, err := api.post(path, jsonUpload)
	logData["httpCode"] = httpCode
	logData["jsonResult"] = jsonResult
	if err == nil && httpCode != http.StatusOK && httpCode != http.StatusCreated {
		err = errors.New("bad response while creating instance")
	}
	if err != nil {
		log.ErrorC("createInstance post", err, logData)
		return
	}
	instance = &models.Instance{}
	if err = json.Unmarshal(jsonResult, instance); err != nil {
		log.ErrorC("createInstance unmarshal", err, logData)
		return
	}
	return
}

// UpdateState tells the Dataset API that the state of a Dataset instance has changed
func (api *DatasetAPI) UpdateInstanceState(instanceID string, newState string) error {
	path := api.url + "/instances/" + instanceID
	logData := log.Data{"URL": path, "new_state": newState}

	jsonUpload, err := json.Marshal(models.Instance{State: newState})
	if err != nil {
		log.ErrorC("updateInstanceState marshal", err, logData)
		return err
	}
	logData["jsonUpload"] = string(jsonUpload)

	jsonResult, httpCode, err := api.put(path, jsonUpload)
	logData["httpCode"] = httpCode
	logData["jsonResult"] = jsonResult
	if err == nil && httpCode != http.StatusOK {
		err = errors.New("bad response while updating instance state")
	}
	if err != nil {
		log.ErrorC("updateInstanceState", err, logData)
		return err
	}
	return nil
}

func (api *DatasetAPI) get(path string, vars url.Values) ([]byte, int, error) {
	return api.callDatasetAPI("GET", path, vars)
}

func (api *DatasetAPI) put(path string, payload []byte) ([]byte, int, error) {
	return api.callDatasetAPI("PUT", path, payload)
}

func (api *DatasetAPI) post(path string, payload []byte) ([]byte, int, error) {
	return api.callDatasetAPI("POST", path, payload)
}

// callDatasetAPI contacts the Dataset API returns the json body (action = PUT, GET, POST, ...)
func (api *DatasetAPI) callDatasetAPI(method, path string, payload interface{}) ([]byte, int, error) {
	logData := log.Data{"URL": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.ErrorC("failed to create url for dataset api call", err, logData)
		return nil, 0, err
	}
	path = URL.String()
	logData["URL"] = path

	var req *http.Request

	if payload != nil && method != "GET" {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload.([]byte)))
		req.Header.Add("Content-type", "application/json")
		logData["payload"] = string(payload.([]byte))
	} else {
		req, err = http.NewRequest(method, path, nil)

		if payload != nil && method == "GET" {
			req.URL.RawQuery = payload.(url.Values).Encode()
			logData["payload"] = payload.(url.Values)
		}
	}
	// check req, above, didn't error
	if err != nil {
		log.ErrorC("failed to create request for dataset api", err, logData)
		return nil, 0, err
	}

	req.Header.Set("Internal-token", api.AuthToken)
	resp, err := api.Client.Do(req)
	if err != nil {
		log.ErrorC("Failed to action dataset api", err, logData)
		return nil, 0, err
	}

	logData["httpCode"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Debug("unexpected status code from api", logData)
	}

	defer resp.Body.Close()
	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC("failed to read body from dataset api", err, logData)
		return nil, resp.StatusCode, err
	}
	return jsonBody, resp.StatusCode, nil
}
