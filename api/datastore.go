package api

import (
	"github.com/ONSdigital/dp-import-api/models"
)
// DataStore - .......
type DataStore interface {
	AddJob(job *models.NewJob) (models.JobInstance, error)
	AddInstance(jobID, instanceID string) (string, error)
	UpdateJobState(jobID string, state *models.JobState) error
	GetInstance(instanceID string) (models.JobInstanceState, error)
	AddUploadedFile(instanceID string, message *models.UploadedFile) error
	AddEvent(instanceID string, event *models.Event) error
	AddDimension(instanceID string, dimension *models.Dimension) error
	GetDimension(instanceID string) ([]models.Dimension, error)
	AddNodeID(instanceID, nodeID string, message *models.Dimension) error
	BuildPublishDatasetMessage(jobID string) (*models.PublishDataset, error)
}
