package api

import (
	"context"
	"encoding/json"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/utils"
	"net/http"
	"strings"

	"github.com/ONSdigital/log.go/log"
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
			log.Event(ctx, "invalid query parameter: offset", log.ERROR, log.Error(err), logData)
			handleErr(ctx, w, err, nil)
			return
		}
	}

	if limitParameter != "" {
		logData["limit"] = limitParameter
		limit, err = utils.ValidatePositiveInt(limitParameter)
		if err != nil {
			log.Event(ctx, "invalid query parameter: limit", log.ERROR, log.Error(err), logData)
			handleErr(ctx, w, err, nil)
			return
		}
	}

	if limit > api.maxLimit {
		logData["max_limit"] = api.maxLimit
		log.Event(ctx, "limit is greater than the maximum allowed", log.ERROR, logData)
		handleCustomErr(ctx, w, errs.ErrorMaximumLimitReached(api.maxLimit), logData, http.StatusBadRequest)
		return
	}

	b, err := api.getJobs(ctx, filterList, offset, limit, logData)
	if err != nil {
		handleErr(ctx, w, err, logData)
		return
	}

	writeResponse(ctx, w, http.StatusOK, b, "getJobs", logData)
	log.Event(ctx, "getJobs endpoint: request successful", logData)
}

func (api *ImportAPI) getJobs(ctx context.Context, filterList []string, offset int, limit int, logData log.Data) (b []byte, err error) {
	jobResults, err := api.dataStore.GetJobs(ctx, filterList, offset, limit)
	if err != nil {
		log.Event(ctx, "getJobs endpoint: failed to retrieve a list of jobs", log.ERROR, log.Error(err), logData)
		return
	}

	b, err = json.Marshal(jobResults)
	if err != nil {
		log.Event(ctx, "getJobs endpoint: failed to marshal jobs resource into bytes", log.ERROR, log.Error(err), logData)
	}

	return b, nil
}
