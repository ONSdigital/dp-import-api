package api

import (
	"encoding/json"
	"net/http"

	errs "github.com/ONSdigital/dp-import-api/apierrors"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

func (api *ImportAPI) increaseProcessedInstanceHandler(w http.ResponseWriter, r *http.Request) {

	defer dphttp.DrainBody(r)

	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["id"]
	instanceID := vars["instance_id"]
	logData := log.Data{jobIDKey: jobID, instanceIDKey: instanceID}

	// Acquire imports lock so that the read and increase are atomic
	lockID, err := api.dataStore.AcquireInstanceLock(ctx, jobID)
	if err != nil {
		handleErr(ctx, w, err, logData)
		return
	}
	defer api.dataStore.UnlockInstance(lockID)

	// Get import job from DB
	job, err := api.dataStore.GetJob(jobID)
	if err != nil {
		handleErr(ctx, w, err, logData)
		return
	}

	// Increase the count for the provided instance
	found := false
	for i, instance := range job.Processed {
		if instance.ID == instanceID {
			job.Processed[i].ProcessedCount++
			found = true
			break
		}
	}

	if !found {
		handleErr(ctx, w, errs.ErrInvalidInstanceID, logData)
		return
	}

	// Update the processedInstance
	if err := api.dataStore.UpdateProcessedInstance(jobID, job.Processed); err != nil {
		handleErr(ctx, w, err, logData)
		return
	}
	log.Info(ctx, "job update completed successfully", logData)

	// marshal full Processed array as a response
	b, err := json.Marshal(job.Processed)
	if err != nil {
		handleErr(ctx, w, err, logData)
		return
	}

	writeResponse(ctx, w, http.StatusOK, b, "increaseProcessedInstanceHandler", logData)
	log.Info(ctx, "created new import job", logData)
}
