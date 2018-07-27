package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

// API aggregates a client and URL and other common data for accessing the API
type API struct {
	Client           common.RCHTTPClienter
	URL              string
	ServiceAuthToken string
}

// CreateInstance tells the Dataset API to create a Dataset instance
func (api *API) CreateInstance(ctx context.Context, job *models.Job, recipeInst *models.RecipeInstance) (instance *models.Instance, err error) {
	path := api.URL + "/instances"
	datasetPath := api.URL + "/datasets/" + recipeInst.DatasetID
	logData := log.Data{"URL": path, "job_id": job.ID, "job_url": job.Links.Self.HRef}

	var jsonUpload []byte
	newInstance := models.CreateInstance(job, recipeInst.DatasetID, datasetPath, recipeInst.CodeLists)

	if jsonUpload, err = json.Marshal(newInstance); err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "CreateInstance: failed to marshal new instance into json"), logData)
		return
	}

	logData["json_request"] = string(jsonUpload)
	log.InfoCtx(ctx, "CreateInstance: create instance request", logData)

	jsonResult, httpCode, err := api.post(ctx, path, jsonUpload)
	logData["http_code"] = httpCode
	logData["json_response"] = string(jsonResult)

	if err == nil && httpCode != http.StatusOK && httpCode != http.StatusCreated {
		err = errors.New("bad response while creating instance")
	}

	if err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "CreateInstance: failed to create instance"), logData)
		return
	}

	instance = &models.Instance{}
	if err = json.Unmarshal(jsonResult, instance); err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "CreateInstance: failed to marshal response to json"), logData)
		return
	}

	return
}

// UpdateInstanceState tells the Dataset API that the state of a Dataset instance has changed
func (api *API) UpdateInstanceState(ctx context.Context, instanceID string, newState string) error {
	path := api.URL + "/instances/" + instanceID
	logData := log.Data{"URL": path, "new_state": newState}

	jsonUpload, err := json.Marshal(models.Instance{State: newState})
	if err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "UpdateInstanceState: failed to marshal instance to json"), logData)
		return err
	}

	logData["json_request"] = string(jsonUpload)
	jsonResult, httpCode, err := api.put(ctx, path, jsonUpload)
	logData["http_code"] = httpCode
	logData["json_response"] = jsonResult

	if err == nil && httpCode != http.StatusOK {
		err = errors.New("bad response while updating instance state")
	}

	if err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "UpdateInstanceState: failed to update instance"), logData)
		return err
	}

	return nil
}

func (api *API) get(ctx context.Context, path string, vars url.Values) ([]byte, int, error) {
	return api.callDatasetAPI(ctx, "GET", path, vars)
}

func (api *API) put(ctx context.Context, path string, payload []byte) ([]byte, int, error) {
	return api.callDatasetAPI(ctx, "PUT", path, payload)
}

func (api *API) post(ctx context.Context, path string, payload []byte) ([]byte, int, error) {
	return api.callDatasetAPI(ctx, "POST", path, payload)
}

// callDatasetAPI contacts the Dataset API returns the json body (action = PUT, GET, POST, ...)
func (api *API) callDatasetAPI(ctx context.Context, method, path string, payload interface{}) ([]byte, int, error) {
	logData := log.Data{"URL": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "callDatasetAPI: failed to create url for dataset api call"), logData)
		return nil, 0, err
	}

	path = URL.String()
	logData["url"] = path

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
		log.ErrorCtx(ctx, errors.Wrap(err, "callDatasetAPI: failed to create request for dataset api"), logData)
		return nil, 0, err
	}

	common.AddAuthHeaders(req.Context(), req, api.ServiceAuthToken)

	resp, err := api.Client.Do(ctx, req)
	if err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "callDatasetAPI: failed to action dataset api"), logData)
		return nil, 0, err
	}

	logData["http_code"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.InfoCtx(ctx, "callDatasetAPI: unexpected status code from api", logData)
	}

	defer resp.Body.Close()
	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "callDatasetAPI: failed to read body from dataset api"), logData)
		return nil, resp.StatusCode, err
	}

	return jsonBody, resp.StatusCode, nil
}
