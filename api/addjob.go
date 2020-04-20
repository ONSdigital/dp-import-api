package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/request"
	"github.com/ONSdigital/log.go/log"
)

func (api *ImportAPI) addJobHandler(w http.ResponseWriter, r *http.Request) {

	defer request.DrainBody(r)

	ctx := r.Context()

	// marshal request body into job structure
	job, err := models.CreateJob(r.Body)
	if err != nil {
		log.Event(ctx, "api endpoint addJob error - Bad client request received", log.ERROR, log.Error(err))

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

	writeResponse(ctx, w, http.StatusCreated, b, "addJob", logData)

	log.Event(ctx, "created new import job", log.INFO, logData)
}

func (api *ImportAPI) addJob(ctx context.Context, job *models.Job, auditParams common.Params, logData log.Data) (b []byte, err error) {
	createdJob, err := api.jobService.CreateJob(ctx, job)
	if err != nil {
		log.Event(ctx, "addJob endpoint: error creating job resource", log.ERROR, log.Error(err), logData)
		return
	}

	logData["job"] = createdJob
	auditParams["createdJobID"] = createdJob.ID

	b, err = json.Marshal(createdJob)
	if err != nil {
		log.Event(ctx, "addJob endpoint: failed to marshal job resource into bytes", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	return
}
