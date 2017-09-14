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

var maxRetries = 5

// DatasetAPI aggreagates a client and URL and other common data for accessing the API
type DatasetAPI struct {
	Client     *rhttp.Client
	URL        string
	MaxRetries int
	AuthToken  string
}

// New creates an DatasetAPI object
func NewDatasetAPI(client *rhttp.Client, datasetAPIURL, datasetAPIAuthToken string) *DatasetAPI {
	return &DatasetAPI{
		Client:     client,
		URL:        datasetAPIURL,
		MaxRetries: maxRetries,
		AuthToken:  datasetAPIAuthToken,
	}
}

func (api *DatasetAPI) GetURL() string {
	return api.URL
}

// CreateInstance tells the Dataset API to create a Dataset instance
func (api *DatasetAPI) CreateInstance(jobID, jobURL string) (instance *models.Instance, err error) {
	path := api.URL + "/instances"
	logData := log.Data{"URL": path, "job_id": jobID, "job_url": jobURL}

	var jsonUpload []byte
	if jsonUpload, err = json.Marshal(models.CreateInstance(jobID, jobURL)); err != nil {
		log.ErrorC("CreateInstance marshal", err, logData)
		return
	}
	logData["jsonUpload"] = jsonUpload
	jsonResult, httpCode, err := api.post(path, jsonUpload)
	logData["httpCode"] = httpCode
	logData["jsonResult"] = jsonResult
	if err == nil && httpCode != http.StatusOK && httpCode != http.StatusCreated {
		err = errors.New("Bad response while creating instance")
	}
	if err != nil {
		log.ErrorC("CreateInstance post", err, logData)
		return
	}
	instance = &models.Instance{}
	if err = json.Unmarshal(jsonResult, instance); err != nil {
		log.ErrorC("CreateInstance unmarshal", err, logData)
		return
	}
	return
}

// UpdateState tells the Dataset API that the state of a Dataset instance has changed
func (api *DatasetAPI) UpdateInstanceState(instanceID string, newState string) error {
	path := api.URL + "/instances/" + instanceID
	logData := log.Data{"URL": path, "new_state": newState}

	jsonUpload, err := json.Marshal(models.Instance{State: newState})
	if err != nil {
		log.ErrorC("UpdateInstanceState marshal", err, logData)
		return err
	}
	logData["jsonUpload"] = string(jsonUpload)

	jsonResult, httpCode, err := api.put(path, jsonUpload)
	logData["httpCode"] = httpCode
	logData["jsonResult"] = jsonResult
	if err == nil && httpCode != http.StatusOK {
		err = errors.New("Bad response while updating instance state")
	}
	if err != nil {
		log.ErrorC("UpdateInstanceState", err, logData)
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
		log.ErrorC("Failed to create URL for DatasetAPI call", err, logData)
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
		log.ErrorC("Failed to create request for DatasetAPI", err, logData)
		return nil, 0, err
	}

	req.Header.Set("Internal-token", api.AuthToken)
	resp, err := api.Client.Do(req)
	if err != nil {
		log.ErrorC("Failed to action DatasetAPI", err, logData)
		return nil, 0, err
	}

	logData["httpCode"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Debug("unexpected status code from API", logData)
	}

	defer resp.Body.Close()
	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC("Failed to read body from DatasetAPI", err, logData)
		return nil, resp.StatusCode, err
	}
	return jsonBody, resp.StatusCode, nil
}
