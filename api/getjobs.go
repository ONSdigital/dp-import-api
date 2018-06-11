package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

func (api *ImportAPI) getJobsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auditParams := common.Params{}
	logData := log.Data{}

	filtersQuery := r.URL.Query().Get("state")
	var filterList []string
	if filtersQuery != "" {
		filterList = strings.Split(filtersQuery, ",")
		logData["filterQuery"] = filtersQuery
		auditParams["filterQuery"] = filtersQuery
	}

	// record attempt to get jobs
	if auditError := api.auditor.Record(ctx, getJobsAction, audit.Attempted, auditParams); auditError != nil {
		handleErr(ctx, w, auditError, logData)
		return
	}

	b, err := api.getJobs(ctx, filterList, auditParams, logData)
	if err != nil {
		// record unsuccessful attempt to get jobs
		if auditError := api.auditor.Record(ctx, getJobsAction, audit.Unsuccessful, auditParams); auditError != nil {
			err = auditError
		}

		handleErr(ctx, w, err, logData)
		return
	}

	// record successful attempt to get jobs
	if apiError := api.auditor.Record(ctx, getJobsAction, audit.Successful, auditParams); apiError != nil {
		handleErr(ctx, w, apiError, logData)
		return
	}

	writeResponse(ctx, w, http.StatusOK, b, "getJobs", logData)

	log.InfoCtx(ctx, "getJobs endpoint: request successful", logData)
}

func (api *ImportAPI) getJobs(ctx context.Context, filterList []string, auditParams common.Params, logData log.Data) (b []byte, err error) {
	jobs, err := api.dataStore.GetJobs(filterList)
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "getJobs endpoint: failed to retrieve a list of jobs"), logData)
		return
	}
	logData["Jobs"] = jobs

	b, err = json.Marshal(jobs)
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "getJobs endpoint: failed to marshal jobs resource into bytes"), logData)
	}

	return
}
