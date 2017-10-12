package job

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/url"
	"github.com/ONSdigital/go-ns/log"
	"github.com/satori/go.uuid"
)

//go:generate moq -out testjob/job_queue.go -pkg testjob . Queue
//go:generate moq -out testjob/dataset_api.go -pkg testjob . DatasetAPI
//go:generate moq -out testjob/recipe_api.go -pkg testjob . RecipeAPI

var ErrInvalidJob = errors.New("the provided Job is not valid")
var ErrGetRecipeFailed = errors.New("failed to get recipe")
var ErrCreateInstanceFailed = errors.New("failed to create a new instance on the dataset api")
var ErrSaveJobFailed = errors.New("failed to save job")

// Service provides job related functionality.
type Service struct {
	dataStore  datastore.DataStorer
	queue      Queue
	datasetAPI DatasetAPI
	recipeAPI  RecipeAPI
	urlBuilder *url.Builder
}

// Queue interface used to queue import jobs.
type Queue interface {
	Queue(job *models.ImportData) error
}

// DatasetAPI interface to the dataset API.
type DatasetAPI interface {
	CreateInstance(ctx context.Context, job *models.Job, recipeInst *models.RecipeInstance) (instance *models.Instance, err error)
	UpdateInstanceState(ctx context.Context, instanceID string, newState string) error
}

// RecipeAPI interface to the recipe API.
type RecipeAPI interface {
	GetRecipe(ctx context.Context, ID string) (*models.Recipe, error)
}

// NewService returns a new instance of a job.Service using the given dependencies.
func NewService(dataStore datastore.DataStorer, queue Queue, datasetAPI DatasetAPI, recipeAPI RecipeAPI, urlBuilder *url.Builder) *Service {
	return &Service{
		dataStore:  dataStore,
		queue:      queue,
		datasetAPI: datasetAPI,
		recipeAPI:  recipeAPI,
		urlBuilder: urlBuilder,
	}
}

// CreateJob creates a new job using the given job instance.
func (service Service) CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {

	if err := job.Validate(); err != nil {
		log.Error(err, log.Data{})
		return nil, ErrInvalidJob
	}

	//Get details needed for instances from Recipe API
	recipe, err := service.recipeAPI.GetRecipe(ctx, job.RecipeID)
	if err != nil {
		log.ErrorC("failed to get recipe details", err, log.Data{"job": job})
		return nil, ErrGetRecipeFailed
	}

	job.ID = (uuid.NewV4()).String()
	job.Links.Self.HRef = service.urlBuilder.GetJobURL(job.ID)

	for _, oi := range recipe.OutputInstances {
		// now create an instance for this file
		instance, err := service.datasetAPI.CreateInstance(ctx, job, &oi)
		if err != nil {
			return nil, ErrCreateInstanceFailed
		}

		job.Links.Instances = append(job.Links.Instances,
			models.IDLink{
				ID:   instance.InstanceID,
				HRef: service.urlBuilder.GetInstanceURL(instance.InstanceID),
			},
		)
	}

	createdJob, err := service.dataStore.AddJob(job)
	if err != nil {
		log.Error(err, log.Data{"job": job})
		return nil, ErrSaveJobFailed
	}

	return createdJob, nil
}

// UpdateJob updates the job for the given jobID with the values in the given job model.
func (service Service) UpdateJob(ctx context.Context, jobID string, job *models.Job) error {

	err := service.dataStore.UpdateJob(jobID, job)
	if err != nil {
		return ErrSaveJobFailed
	}

	log.Info("job updated", log.Data{"job": job, "job_id": jobID})
	if job.State == "submitted" {
		tasks, err := service.prepareJob(ctx, jobID)
		if err != nil {
			log.Error(err, log.Data{"jobState": job, "job_id": jobID})
			return err
		}

		err = service.queue.Queue(tasks)
		if err != nil {
			log.Error(err, log.Data{"tasks": tasks})
			return err
		}

		log.Info("import job was queued", log.Data{"job": job, "job_id": jobID})
	}

	return nil
}

// PrepareJob returns a format ready to send to downstream services via kafka
func (service Service) prepareJob(ctx context.Context, jobID string) (*models.ImportData, error) {

	importJob, err := service.dataStore.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	recipe, err := service.recipeAPI.GetRecipe(ctx, importJob.RecipeID)
	if err != nil {
		return nil, err
	}

	var instanceIds []string
	for _, instanceRef := range importJob.Links.Instances {
		instanceIds = append(instanceIds, instanceRef.ID)

		if err = service.datasetAPI.UpdateInstanceState(ctx, instanceRef.ID, "submitted"); err != nil {
			return nil, err
		}
	}

	return &models.ImportData{
		JobID:         jobID,
		Recipe:        importJob.RecipeID,
		Format:        recipe.Format,
		UploadedFiles: importJob.UploadedFiles,
		InstanceIDs:   instanceIds,
	}, nil
}
