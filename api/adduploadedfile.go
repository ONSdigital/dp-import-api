package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-import-api/models"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

func (api *ImportAPI) addUploadedFileHandler(w http.ResponseWriter, r *http.Request) {

	defer dphttp.DrainBody(r)

	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}

	uploadedFile, err := models.CreateUploadedFile(r.Body)
	if err != nil {
		log.Event(ctx, "addUploadFile endpoint: failed to create uploaded file resource", log.ERROR, log.Error(err), logData)
		handleErr(ctx, w, err, nil)
		return
	}

	logData["file"] = uploadedFile

	if err := api.addUploadFile(ctx, uploadedFile, jobID, logData); err != nil {
		handleErr(ctx, w, err, logData)
		return
	}

	log.Event(ctx, "added uxploaded file to job", logData)
}

func (api *ImportAPI) addUploadFile(ctx context.Context, uploadedFile *models.UploadedFile, jobID string, logData log.Data) (err error) {

	if err = api.dataStore.AddUploadedFile(jobID, uploadedFile); err != nil {
		log.Event(ctx, "addUploadFile endpoint: failed to store uploaded file resource", log.ERROR, log.Error(err), logData)
	}

	return
}
