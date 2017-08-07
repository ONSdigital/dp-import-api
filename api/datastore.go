package api

import (
	"github.com/ONSdigital/dp-import-api/models"
)

// DataStore is a interface used to store instances and jobs
type DataStore interface {
	AddJob(host string, job *models.Job) (models.Job, error)
	GetJob(host string, jobID string) (models.Job, error)
	GetJobs(host string, filters []string) ([]models.Job, error)
	UpdateJobState(jobID string, state *models.Job, withOutRestrictions bool) error
	GetInstance(host, instanceID string) (models.Instance, error)
	GetInstances(host string, filter []string) ([]models.Instance, error)
	UpdateInstance(instanceID string, instance *models.Instance) error
	AddUploadedFile(instanceID string, message *models.UploadedFile) error
	AddEvent(instanceID string, event *models.Event) error
	AddDimension(instanceID string, dimension *models.Dimension) error
	GetDimensions(instanceID string) ([]models.Dimension, error)
	GetDimensionValues(instanceID, dimensionName string) (models.UniqueDimensionValues, error)
	AddNodeID(instanceID, nodeID string, message *models.Dimension) error
	UpdateObservationCount(instanceID string, count int) error
	PrepareImportJob(jobID string) (*models.ImportData, error)
}
