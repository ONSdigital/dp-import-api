package recipe

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"fmt"

	"github.com/ONSdigital/dp-import-api/models"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/pkg/errors"
)

// API provides a client for calling the Recipe API.
type API struct {
	Client dphttp.Clienter
	URL    string
}

// GetRecipe from an ID
func (api *API) GetRecipe(ctx context.Context, ID string) (*models.Recipe, error) {

	logData := log.Data{"ID": ID}

	jsonResult, httpCode, err := api.get(ctx, fmt.Sprintf("%s/recipes/%s", api.URL, ID), nil)
	logData["http_code"] = httpCode

	if err == nil && httpCode != http.StatusOK {
		recipeErr := errors.New("bad response")
		log.Event(ctx, "GetRecipe: failed to retrieve recipe", log.ERROR, log.Error(recipeErr), logData)
		return nil, recipeErr
	}

	if err != nil {
		log.Event(ctx, "GetRecipe: failed to retrieve recipe", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	var recipe *models.Recipe
	if err = json.Unmarshal(jsonResult, &recipe); err != nil {
		logData["json_response"] = jsonResult
		log.Event(ctx, "GetRecipe: failed to unmarshal json response from the recipe api", log.ERROR, log.Error(err), logData)
	}

	return recipe, nil
}

func (api *API) get(ctx context.Context, path string, vars url.Values) ([]byte, int, error) {
	return api.callRecipeAPI(ctx, "GET", path, vars)
}

// callRecipeAPI contacts the Dataset API returns the json body (action = PUT, GET, POST, ...)
func (api *API) callRecipeAPI(ctx context.Context, method, path string, payload interface{}) ([]byte, int, error) {
	logData := log.Data{"path": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.Event(ctx, "callRecipeAPI: failed to create URL for recipe api call", log.ERROR, log.Error(err), logData)
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
		log.Event(ctx, "callRecipeAPI: failed to create request for recipe api", log.ERROR, log.Error(err), logData)
		return nil, 0, err
	}

	resp, err := api.Client.Do(ctx, req)
	if err != nil {
		log.Event(ctx, "callRecipeAPI: failed to action recipe api", log.ERROR, log.Error(err), logData)
		return nil, 0, err
	}

	logData["http_code"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Event(ctx, "callRecipeAPI: unexpected status code from API", log.INFO, logData)
	}

	defer resp.Body.Close()
	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Event(ctx, "callRecipeAPI: failed to read body from recipe api", log.ERROR, log.Error(err), logData)
		return nil, resp.StatusCode, err
	}

	return jsonBody, resp.StatusCode, nil
}
