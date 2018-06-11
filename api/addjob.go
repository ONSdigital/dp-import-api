package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

func (api *ImportAPI) addJobHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// record attempt to add a job
	if auditError := api.auditor.Record(ctx, addJobAction, audit.Attempted, nil); auditError != nil {
		handleErr(ctx, w, auditError, nil)
		return
	}

	// marshal request body into job structure
	job, err := models.CreateJob(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "api endpoint addJob error - Bad client request received"), nil)

		// record failure to add job
		if auditError := api.auditor.Record(ctx, addJobAction, audit.Unsuccessful, nil); auditError != nil {
			err = auditError
		}

		handleErr(ctx, w, err, nil)
		return
	}

	logData := log.Data{"recipeID": job.RecipeID}
	auditParams := common.Params{"recipeID": job.RecipeID}

	b, err := api.addJob(ctx, job, auditParams, logData)
	if err != nil {
		// record unsuccessful attempt to add job
		if auditError := api.auditor.Record(ctx, addJobAction, audit.Unsuccessful, auditParams); auditError != nil {
			err = auditError
		}

		handleErr(ctx, w, err, logData)
		return
	}
	// record successful attempt to add job
	api.auditor.Record(ctx, addJobAction, audit.Successful, auditParams)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeBody(ctx, w, b, "addJob", logData)

	log.InfoCtx(ctx, "created new import job", logData)
}

func (api *ImportAPI) addJob(ctx context.Context, job *models.Job, auditParams common.Params, logData log.Data) (b []byte, err error) {
	createdJob, err := api.jobService.CreateJob(ctx, job)
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "addJob endpoint: error creating job resource"), logData)
		return
	}

	logData["job"] = createdJob
	auditParams["createdJobID"] = createdJob.ID

	b, err = json.Marshal(createdJob)
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "addJob endpoint: failed to marshal job resource into bytes"), logData)
		return nil, err
	}

	return
}
