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

func (ds *DataStore) AddJob(importJob *models.NewJob) (models.JobInstance, error) {
	if ds.InternalError {
		return models.JobInstance{}, internalError
	}
	return models.JobInstance{JobID: "34534543543"}, nil
}

func (ds *DataStore) AddInstance(joId, dataset string) (string, error) {
	if ds.NotFound {
		return "", utils.JobNotFoundError
	}
	if ds.InternalError {
		return "", internalError
	}
	return "123", nil
}

func (ds *DataStore) UpdateJobState(string, *models.JobState) error {
	if ds.NotFound {
		return utils.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) GetInstance(instanceId string) (models.JobInstanceState, error) {
	if ds.NotFound {
		return models.JobInstanceState{}, utils.JobNotFoundError
	}
	if ds.InternalError {
		return models.JobInstanceState{}, internalError
	}
	return models.JobInstanceState{InstanceID: "234234", Dataset: "123", State: "Created",
		Events:                                []models.Event{models.Event{Type: "Info", Message: "Create at ...", Time: "00000", MessageOffset: "0"}}}, nil
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
	return []models.Dimension{models.Dimension{NodeName: "1234-geography.newport", Value: "newport", NodeID: "234234234"}}, nil
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
	return nil, nil
}
