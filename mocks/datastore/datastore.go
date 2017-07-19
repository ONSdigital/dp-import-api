package mocks

import (
	"fmt"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/utils"
)

var internalError = fmt.Errorf("DataStore internal error")

type DataStore struct {
	NotFound      bool
	InternalError bool
}

func (ds *DataStore) AddJob(importJob *models.Job) (models.Job, error) {
	if ds.InternalError {
		return models.Job{}, internalError
	}
	return models.Job{JobID: "34534543543"}, nil
}

func (ds *DataStore) AddInstance(joId string) (string, error) {
	if ds.NotFound {
		return "", utils.JobNotFoundError
	}
	if ds.InternalError {
		return "", internalError
	}
	return "123", nil
}

func (ds *DataStore) UpdateJobState(string, *models.Job) error {
	if ds.NotFound {
		return utils.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) GetInstance(instanceId string) (models.Instance, error) {
	if ds.NotFound {
		return models.Instance{}, utils.JobNotFoundError
	}
	if ds.InternalError {
		return models.Instance{}, internalError
	}
	return models.Instance{InstanceID: "234234", State: "Created",
		Events: []models.Event{models.Event{Type: "Info", Message: "Create at ...", Time: "00000", MessageOffset: "0"}}}, nil
}

func (ds *DataStore) AddUploadedFile(instanceId string, s3file *models.UploadedFile) error {
	if ds.NotFound {
		return utils.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) AddEvent(instanceId string, event *models.Event) error {
	if ds.NotFound {
		return utils.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) AddDimension(instanceId string, dimension *models.Dimension) error {
	if ds.NotFound {
		return utils.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) GetDimension(instanceId string) ([]models.Dimension, error) {
	if ds.NotFound {
		return []models.Dimension{}, utils.JobNotFoundError
	}
	if ds.InternalError {
		return []models.Dimension{}, internalError
	}
	return []models.Dimension{models.Dimension{Name: "1234-geography.newport", Value: "newport", NodeID: "234234234"}}, nil
}

func (ds *DataStore) AddNodeID(instanceId, nodeId string, message *models.Dimension) error {
	if ds.NotFound {
		return utils.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) BuildPublishDatasetMessage(jobId string) (*models.PublishDataset, error) {
	if ds.NotFound {
		return nil, utils.JobNotFoundError
	}
	if ds.InternalError {
		return nil, internalError
	}
	return &models.PublishDataset{Recipe: "test", InstanceIds: []string{"1", "2", "3"},
		UploadedFiles: []models.UploadedFile{models.UploadedFile{URL: "s3//aws/bucket/file.xls", AliasName: "test"}}}, nil
}