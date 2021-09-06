package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-import-api/models"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

func (api *ImportAPI) updateJobHandler(w http.ResponseWriter, r *http.Request) {

	defer dphttp.DrainBody(r)

	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}

	if err := api.updateJob(ctx, r, jobID, logData); err != nil {
		handleErr(ctx, w, err, logData)
		return
	}
	log.Info(ctx, "job update successful", logData)
}

func (api *ImportAPI) updateJob(ctx context.Context, r *http.Request, jobID string, logData log.Data) (err error) {

	job, err := models.CreateJob(r.Body)
	if err != nil {
		log.Error(ctx, "updateJob endpoint: failed to update job resource", err, logData)
		return
	}
	logData["job"] = job

	if err = job.ValidateState(); err != nil {
		logData["state"] = job.State
		log.Error(ctx, "updateJob endpoint: failed to store updated job resource", err, logData)
		return
	}

	if err = api.jobService.UpdateJob(ctx, jobID, job); err != nil {
		log.Error(ctx, "updateJob endpoint: failed to store updated job resource", err, logData)
	}

	return
}
