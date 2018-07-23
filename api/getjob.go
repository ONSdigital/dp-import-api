package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (api *ImportAPI) getJobHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}
	auditParams := common.Params{jobIDKey: jobID}

	b, err := api.getJob(ctx, jobID, auditParams, logData)
	if err != nil {
		// record unsuccessful attempt to get jobs
		if auditError := api.auditor.Record(ctx, getJobAction, audit.Unsuccessful, auditParams); auditError != nil {
			err = auditError
		}

		handleErr(ctx, w, err, logData)
		return
	}

	// record successful attempt to get jobs
	if auditError := api.auditor.Record(ctx, getJobAction, audit.Successful, auditParams); auditError != nil {
		handleErr(ctx, w, auditError, logData)
		return
	}

	writeResponse(ctx, w, http.StatusOK, b, "getJob", logData)

	log.InfoCtx(ctx, "getJob endpoint: request successful", logData)
}

func (api *ImportAPI) getJob(ctx context.Context, jobID string, auditParams common.Params, logData log.Data) (b []byte, err error) {
	job, err := api.dataStore.GetJob(jobID)
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "getJob endpoint: failed to find job"), logData)
		return
	}

	logData["job"] = job

	b, err = json.Marshal(job)
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "getJob endpoint: failed to marshal jobs resource into bytes"), logData)
	}
	return
}
