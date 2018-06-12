package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (api *ImportAPI) addUploadedFileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["id"]
	logData := log.Data{jobIDKey: jobID}
	auditParams := common.Params{jobIDKey: jobID}

	// record attempt to add uploaded file to job
	if auditError := api.auditor.Record(ctx, uploadFileAction, audit.Attempted, auditParams); auditError != nil {
		handleErr(ctx, w, auditError, logData)
		return
	}

	uploadedFile, err := models.CreateUploadedFile(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "addUploadFile endpoint: failed to create uploaded file resource"), logData)

		// record failure to add uploaded file
		if auditError := api.auditor.Record(ctx, uploadFileAction, audit.Unsuccessful, auditParams); auditError != nil {
			err = auditError
		}

		handleErr(ctx, w, err, nil)
		return
	}

	logData["file"] = uploadedFile
	auditParams["fileAlias"] = uploadedFile.AliasName
	auditParams["fileURL"] = uploadedFile.URL

	if err := api.addUploadFile(ctx, uploadedFile, jobID, auditParams, logData); err != nil {
		// record unsuccessful attempt to add uploaded file to job
		if auditError := api.auditor.Record(ctx, uploadFileAction, audit.Unsuccessful, auditParams); auditError != nil {
			err = auditError
		}

		handleErr(ctx, w, err, logData)
		return
	}

	// record successful attempt to add uploaded file to job
	api.auditor.Record(ctx, uploadFileAction, audit.Successful, auditParams)

	log.InfoCtx(ctx, "added uxploaded file to job", logData)
}

func (api *ImportAPI) addUploadFile(ctx context.Context, uploadedFile *models.UploadedFile, jobID string, auditParams common.Params, logData log.Data) (err error) {

	if err = api.dataStore.AddUploadedFile(jobID, uploadedFile); err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "addUploadFile endpoint: failed to store uploaded file resource"), logData)
	}

	return
}
