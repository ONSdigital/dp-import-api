package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

func (api *ImportAPI) getJobHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}

	b, err := api.getJob(ctx, jobID, logData)
	if err != nil {
		handleErr(ctx, w, err, logData)
		return
	}

	writeResponse(ctx, w, http.StatusOK, b, "getJob", logData)
	log.Event(ctx, "getJob endpoint: request successful", logData)
}

func (api *ImportAPI) getJob(ctx context.Context, jobID string, logData log.Data) (b []byte, err error) {
	job, err := api.dataStore.GetJob(jobID)
	if err != nil {
		log.Event(ctx, "getJob endpoint: failed to find job", log.ERROR, log.Error(err), logData)
		return
	}

	logData["job"] = job

	b, err = json.Marshal(job)
	if err != nil {
		log.Event(ctx, "getJob endpoint: failed to marshal jobs resource into bytes", log.ERROR, log.Error(err), logData)
	}
	return
}
