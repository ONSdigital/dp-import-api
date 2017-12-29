package mongo

import (
	"errors"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/models"
)

var internalError = errors.New("DataStore internal error")

type DataStorer struct {
	NotFound      bool
	InternalError bool
}

func (ds *DataStorer) AddJob(importJob *models.Job) (*models.Job, error) {
	if ds.InternalError {
		return &models.Job{}, internalError
	}
	return &models.Job{ID: "34534543543"}, nil
}

func (ds *DataStorer) GetJobs(filter []string) ([]models.Job, error) {
	if ds.InternalError {
		return []models.Job{}, internalError
	}
	return []models.Job{{ID: "34534543543"}}, nil
}

func (ds *DataStorer) GetJob(jobID string) (*models.Job, error) {
	if ds.InternalError {
		return &models.Job{}, internalError
	}
	if ds.NotFound {
		return &models.Job{}, api_errors.JobNotFoundError
	}
	return &models.Job{ID: "34534543543"}, nil
}

func (ds *DataStorer) AddInstance(jobID string) (string, error) {
	if ds.NotFound {
		return "", api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return "", internalError
	}
	return "123", nil
}

func (ds *DataStorer) UpdateJob(string, *models.Job) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStorer) UpdateJobState(string, string) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStorer) AddUploadedFile(jobID string, message *models.UploadedFile) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStorer) UpdateInstanceState(jobID, instanceID, taskID, newState string) (err error) {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStorer) UpdateInstanceTaskState(jobID, instanceID, taskID, newState string) (err error) {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}
