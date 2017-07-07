package mocks

import (
	"fmt"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
)

var internalError = fmt.Errorf("DataStore internal error")

// Notes on why its being mocked
type DataStore struct {
	NotFound      bool
	InternalError bool
}

func (ds *DataStore) AddJob(importJob *models.ImportJob) (models.JobInstance, error) {
	if ds.InternalError {
		return models.JobInstance{}, internalError
	}
	return models.JobInstance{JobId: "34534543543"}, nil
}

func (ds *DataStore) AddInstance(dataset string) (string, error) {
	if ds.NotFound {
		return "", datastore.JobNotFoundError
	}
	if ds.InternalError {
		return "", internalError
	}
	return "123", nil
}

func (ds *DataStore) UpdateJobState(string, *models.JobState) error {
	if ds.NotFound {
		return datastore.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return  nil
}

func (ds *DataStore) GetInstance(instanceId string) (models.ImportJobState, error) {
	if ds.NotFound {
		return models.ImportJobState{}, datastore.JobNotFoundError
	}
	if ds.InternalError {
		return models.ImportJobState{}, internalError
	}
	return models.ImportJobState{InstanceId: "234234", Dataset: "123", State: "Created",
		Events: []models.Event{models.Event{Type: "Info", Message: "Create at ...", Time: "00000", MessageOffset: "0"}}}, nil
}

func (ds *DataStore) AddS3File(instanceId string, s3file *models.S3File) error {
	if ds.NotFound {
		return datastore.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) AddEvent(instanceId string, event *models.Event) error {
	if ds.NotFound {
		return datastore.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) AddDimension(instanceId string, dimension *models.Dimension) error {
	if ds.NotFound {
		return datastore.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}

func (ds *DataStore) GetDimension(instanceId string) ([]models.Dimension, error) {
	if ds.NotFound {
		return []models.Dimension{}, datastore.JobNotFoundError
	}
	if ds.InternalError {
		return []models.Dimension{}, internalError
	}
	return []models.Dimension{models.Dimension{NodeName: "1234-geography.newport", Value: "newport", NodeId: "234234234"}}, nil
}

func (ds *DataStore) AddNodeId(instanceId, nodeId string, message *models.Dimension) error {
	if ds.NotFound {
		return datastore.JobNotFoundError
	}
	if ds.InternalError {
		return internalError
	}
	return nil
}
