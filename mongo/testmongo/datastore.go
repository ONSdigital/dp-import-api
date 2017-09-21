package mongo

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/dataset/interface"
	"github.com/ONSdigital/dp-import-api/models"
)

var internalError = errors.New("DataStore internal error")

type DataStorer struct {
	NotFound      bool
	InternalError bool
}

func (ds *DataStorer) AddJob(ctx context.Context, importJob *models.Job, selfURL string, datasetAPI dataset.DatasetAPIer) (*models.Job, error) {
	if ds.InternalError {
		return &models.Job{}, internalError
	}
	return &models.Job{JobID: "34534543543"}, nil
}

func (ds *DataStorer) GetJobs(filter []string) ([]models.Job, error) {
	if ds.InternalError {
		return []models.Job{}, internalError
	}
	return []models.Job{models.Job{JobID: "34534543543"}}, nil
}

func (ds *DataStorer) GetJob(jobID string) (*models.Job, error) {
	if ds.InternalError {
		return &models.Job{}, internalError
	}
	if ds.NotFound {
		return &models.Job{}, api_errors.JobNotFoundError
	}
	return &models.Job{JobID: "34534543543"}, nil
}

func (ds *DataStorer) AddInstance(joID string) (string, error) {
	if ds.NotFound {
		return "", api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return "", internalError
	}
	return "123", nil
}

func (ds *DataStorer) UpdateJob(string, *models.Job, bool) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStorer) UpdateJobState(string, string, bool) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStorer) AddUploadedFile(ctx context.Context, instanceID string, message *models.UploadedFile, datasetAPI dataset.DatasetAPIer, selfURL string) (*models.Instance, error) {
	if ds.NotFound {
		return nil, api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return nil, internalError
	}
	return nil, nil
}

func (ds *DataStorer) PrepareJob(ctx context.Context, dataset dataset.DatasetAPIer, jobID string) (*models.ImportData, error) {
	if ds.NotFound {
		return nil, api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return nil, internalError
	}
	return &models.ImportData{Recipe: "test", InstanceIDs: []string{"1", "2", "3"},
		UploadedFiles: &[]models.UploadedFile{models.UploadedFile{URL: "s3//aws/bucket/file.xls", AliasName: "test"}}}, nil
}

func (ds *DataStorer) UpdateObservationCount(instanceID string, count int) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}
