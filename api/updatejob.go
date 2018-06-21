package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/request"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (api *ImportAPI) updateJobHandler(w http.ResponseWriter, r *http.Request) {

	defer request.DrainBody(r)

	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}
	auditParams := common.Params{jobIDKey: jobID}

	// record attempt to update job
	if auditError := api.auditor.Record(ctx, updateJobAction, audit.Attempted, auditParams); auditError != nil {
		handleErr(ctx, w, auditError, logData)
		return
	}

	if err := api.updateJob(ctx, r, jobID, auditParams, logData); err != nil {
		// record unsuccessful attempt to update job
		if auditError := api.auditor.Record(ctx, updateJobAction, audit.Unsuccessful, auditParams); auditError != nil {
			err = auditError
		}

		handleErr(ctx, w, err, logData)
		return
	}

	// record successful attempt to update job
	api.auditor.Record(ctx, updateJobAction, audit.Successful, auditParams)

	audit.LogInfo(ctx, "job update completed successfully", logData)
}

func (api *ImportAPI) updateJob(ctx context.Context, r *http.Request, jobID string, auditParams common.Params, logData log.Data) (err error) {

	job, err := models.CreateJob(r.Body)
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "updateJob endpoint: failed to update job resource"), logData)
		return
	}
	logData["job"] = job

	if err = job.ValidateState(); err != nil {
		logData["state"] = job.State
		log.ErrorCtx(ctx, errors.WithMessage(err, "updateJob endpoint: failed to store updated job resource"), logData)
		return
	}

	if err = api.jobService.UpdateJob(ctx, jobID, job); err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "updateJob endpoint: failed to store updated job resource"), logData)
	}

	return
}
