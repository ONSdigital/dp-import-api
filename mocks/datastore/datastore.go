package mocks

import (
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/api-errors"
	"errors"
)

var internalError = errors.New("DataStore internal error")

type DataStore struct {
	NotFound      bool
	InternalError bool
}

func (ds *DataStore) AddJob(host string, importJob *models.Job) (models.Job, error) {
	if ds.InternalError {
		return models.Job{}, internalError
	}
	return models.Job{JobID: "34534543543"}, nil
}

func (ds *DataStore) AddInstance(joID string) (string, error) {
	if ds.NotFound {
		return "", api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return "", internalError
	}
	return "123", nil
}

func (ds *DataStore) UpdateJobState(string, *models.Job) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) GetInstance(instanceID string) (models.Instance, error) {
	if ds.NotFound {
		return models.Instance{}, api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return models.Instance{}, internalError
	}
	return models.Instance{InstanceID: "234234", State: "Created"}, nil
}

func (ds *DataStore) UpdateInstance(instanceID string, instance *models.Instance) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) AddUploadedFile(instanceID string, s3file *models.UploadedFile) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) AddEvent(instanceID string, event *models.Event) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) AddDimension(instanceID string, dimension *models.Dimension) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) GetDimension(instanceID string) ([]models.Dimension, error) {
	if ds.NotFound {
		return []models.Dimension{}, api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return []models.Dimension{}, internalError
	}
	return []models.Dimension{models.Dimension{Name: "1234-geography.newport", Value: "newport", NodeID: "234234234"}}, nil
}

func (ds *DataStore) AddNodeID(instanceID, nodeID string, message *models.Dimension) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) BuildImportDataMessage(jobID string) (*models.ImportData, error) {
	if ds.NotFound {
		return nil, api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return nil, internalError
	}
	return &models.ImportData{Recipe: "test", InstanceIDs: []string{"1", "2", "3"},
		UploadedFiles: []models.UploadedFile{models.UploadedFile{URL: "s3//aws/bucket/file.xls", AliasName: "test"}}}, nil
}
