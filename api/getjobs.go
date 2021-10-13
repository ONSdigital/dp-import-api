package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/utils"

	"github.com/ONSdigital/log.go/v2/log"
)

func (api *ImportAPI) getJobsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logData := log.Data{}

	filtersQuery := r.URL.Query().Get("state")
	var filterList []string
	if filtersQuery != "" {
		filterList = strings.Split(filtersQuery, ",")
		logData["filterQuery"] = filtersQuery
	}

	offsetParameter := r.URL.Query().Get("offset")
	limitParameter := r.URL.Query().Get("limit")

	limit := api.defaultLimit
	offset := api.defaultOffset

	var err error

	if offsetParameter != "" {
		logData["offset"] = offsetParameter
		offset, err = utils.ValidatePositiveInt(offsetParameter)
		if err != nil {
			log.Error(ctx, "invalid query parameter: offset", err, logData)
			handleErr(ctx, w, err, nil)
			return
		}
	}

	if limitParameter != "" {
		logData["limit"] = limitParameter
		limit, err = utils.ValidatePositiveInt(limitParameter)
		if err != nil {
			log.Error(ctx, "invalid query parameter: limit", err, logData)
			handleErr(ctx, w, err, nil)
			return
		}
	}

	if limit > api.maxLimit {
		logData["max_limit"] = api.maxLimit
		err = errs.ErrorMaximumLimitReached(api.maxLimit)
		log.Error(ctx, "limit is greater than the maximum allowed", err, logData)
		handleCustomErr(ctx, w, err, logData, http.StatusBadRequest)
		return
	}

	b, err := api.getJobs(ctx, filterList, offset, limit, logData)
	if err != nil {
		handleErr(ctx, w, err, logData)
		return
	}

	writeResponse(ctx, w, http.StatusOK, b, "getJobs", logData)
	log.Info(ctx, "getJobs endpoint: request successful", logData)
}

func (api *ImportAPI) getJobs(ctx context.Context, filterList []string, offset int, limit int, logData log.Data) (b []byte, err error) {
	jobResults, err := api.dataStore.GetJobs(ctx, filterList, offset, limit)
	if err != nil {
		log.Error(ctx, "getJobs endpoint: failed to retrieve a list of jobs", err, logData)
		return
	}

	b, err = json.Marshal(jobResults)
	if err != nil {
		log.Error(ctx, "getJobs endpoint: failed to marshal jobs resource into bytes", err, logData)
	}

	return b, nil
}
