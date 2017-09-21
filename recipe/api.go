package recipe

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"fmt"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rhttp"
)

// API provides a client for calling the Recipe API.
type API struct {
	Client *rhttp.Client
	URL    string
}

// NewAPI returns a new API instance.
func NewAPI(client *rhttp.Client, url string) *API {
	return &API{
		Client: client,
		URL: url,
	}
}

// GetRecipe from an ID
func (api *API) GetRecipe(ID string) (*models.Recipe, error) {

	logData := log.Data{"ID": ID}

	jsonResult, httpCode, err := api.get(fmt.Sprintf("%s/recipes/%s", api.URL, ID), nil)
	logData["httpCode"] = httpCode
	logData["jsonResult"] = jsonResult

	if err == nil && httpCode != http.StatusOK {
		return nil, errors.New("Bad response while creating instance")
	}

	if err != nil {
		log.ErrorC("CreateInstance post", err, logData)
		return nil, err
	}

	var recipe *models.Recipe
	json.Unmarshal(jsonResult, &recipe)
	return recipe, nil
}

func (api *API) get(path string, vars url.Values) ([]byte, int, error) {
	return api.callRecipeAPI("GET", path, vars)
}

// callRecipeAPI contacts the Dataset API returns the json body (action = PUT, GET, POST, ...)
func (api *API) callRecipeAPI(method, path string, payload interface{}) ([]byte, int, error) {
	logData := log.Data{"URL": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.ErrorC("failed to create URL for recipe api call", err, logData)
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
		log.ErrorC("failed to create request for recipe api", err, logData)
		return nil, 0, err
	}

	resp, err := api.Client.Do(req)
	if err != nil {
		log.ErrorC("failed to action recipe api", err, logData)
		return nil, 0, err
	}

	logData["httpCode"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Debug("unexpected status code from API", logData)
	}

	defer resp.Body.Close()
	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC("failed to read body from recipe api", err, logData)
		return nil, resp.StatusCode, err
	}
	return jsonBody, resp.StatusCode, nil
}
