package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-import-api/models"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
)

func (api *ImportAPI) addJobHandler(w http.ResponseWriter, r *http.Request) {

	defer dphttp.DrainBody(r)

	ctx := r.Context()

	// marshal request body into job structure
	job, err := models.CreateJob(r.Body)
	if err != nil {
		log.Error(ctx, "api endpoint addJob error - Bad client request received", err)
		handleErr(ctx, w, err, nil)
		return
	}

	logData := log.Data{"recipeID": job.RecipeID}

	b, err := api.addJob(ctx, job, logData)
	if err != nil {
		handleErr(ctx, w, err, logData)
		return
	}

	writeResponse(ctx, w, http.StatusCreated, b, "addJob", logData)
	log.Info(ctx, "created new import job", logData)
}

func (api *ImportAPI) addJob(ctx context.Context, job *models.Job, logData log.Data) (b []byte, err error) {
	createdJob, err := api.jobService.CreateJob(ctx, job)
	if err != nil {
		log.Error(ctx, "addJob endpoint: error creating job resource", err, logData)
		return
	}

	logData["job"] = createdJob

	b, err = json.Marshal(createdJob)
	if err != nil {
		log.Error(ctx, "addJob endpoint: failed to marshal job resource into bytes", err, logData)
		return nil, err
	}

	return
}
