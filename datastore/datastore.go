package datastore

import (
	"errors"
	"github.com/ONSdigital/dp-import-api/models"
)

type DataStore interface {
	AddJob(*models.ImportJob) (models.JobInstance, error)
	AddInstance(string) (string, error)
	UpdateJobState(string, *models.JobState) error
	GetInstance(instanceId string) (models.ImportJobState, error)
	AddS3File(instanceId string, message *models.S3File) error
	AddEvent(instanceId string, event *models.Event) error
	AddDimension(instanceId string, dimension *models.Dimension) error
	GetDimension(instanceId string) ([]models.Dimension, error)
	AddNodeId(instanceId, nodeId string, message *models.Dimension) error
}

var JobNotFoundError = errors.New("No job instance found")
