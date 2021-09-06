package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-import-api/models"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
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
		log.Error(ctx, "addUploadFile endpoint: failed to create uploaded file resource", err, logData)
		handleErr(ctx, w, err, nil)
		return
	}

	logData["file"] = uploadedFile

	if err := api.addUploadFile(ctx, uploadedFile, jobID, logData); err != nil {
		handleErr(ctx, w, err, logData)
		return
	}

	log.Info(ctx, "added uploaded file to job", logData)
}

func (api *ImportAPI) addUploadFile(ctx context.Context, uploadedFile *models.UploadedFile, jobID string, logData log.Data) (err error) {

	if err = api.dataStore.AddUploadedFile(jobID, uploadedFile); err != nil {
		log.Error(ctx, "addUploadFile endpoint: failed to store uploaded file resource", err, logData)
	}

	return
}
