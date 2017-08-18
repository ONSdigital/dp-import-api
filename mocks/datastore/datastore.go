package mocks

import (
	"errors"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/models"
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

func (ds *DataStore) GetJobs(host string, filter []string) ([]models.Job, error) {
	if ds.InternalError {
		return []models.Job{}, internalError
	}
	return []models.Job{models.Job{JobID: "34534543543"}}, nil
}

func (ds *DataStore) GetJob(host string, jobID string) (models.Job, error) {
	if ds.InternalError {
		return models.Job{}, internalError
	}
	if ds.NotFound {
		return models.Job{}, api_errors.JobNotFoundError
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

func (ds *DataStore) UpdateJobState(string, *models.Job, bool) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) GetInstance(host, instanceID string) (models.Instance, error) {
	if ds.NotFound {
		return models.Instance{}, api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return models.Instance{}, internalError
	}
	return models.Instance{InstanceID: "234234", State: "Created"}, nil
}

func (ds *DataStore) GetInstances(host string, filter []string) ([]models.Instance, error) {

	if ds.InternalError {
		return []models.Instance{}, internalError
	}
	return []models.Instance{models.Instance{InstanceID: "234234", State: "Created"}}, nil
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

func (ds *DataStore) GetDimensions(instanceID string) ([]models.Dimension, error) {
	if ds.NotFound {
		return []models.Dimension{}, api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return []models.Dimension{}, internalError
	}
	return []models.Dimension{models.Dimension{Name: "1234-geography.newport", Value: "newport", NodeID: "234234234"}}, nil
}

func (ds *DataStore) GetDimensionValues(instanceID, dimensionName string) (*models.UniqueDimensionValues, error) {
	if ds.NotFound {
		return nil, api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return nil, internalError
	}
	return &models.UniqueDimensionValues{Name: dimensionName, Values: []string{"123", "321"}}, nil
}

func (ds *DataStore) AddNodeID(instanceID string, dimension *models.Dimension) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) PrepareImportJob(jobID string) (*models.ImportData, error) {
	if ds.NotFound {
		return nil, api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return nil, internalError
	}
	return &models.ImportData{Recipe: "test", InstanceIDs: []string{"1", "2", "3"},
		UploadedFiles: []models.UploadedFile{models.UploadedFile{URL: "s3//aws/bucket/file.xls", AliasName: "test"}}}, nil
}

func (ds *DataStore) UpdateObservationCount(instanceID string, count int) error {
	if ds.NotFound {
		return api_errors.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}
