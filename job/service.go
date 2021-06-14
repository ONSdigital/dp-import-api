package job

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/url"
	"github.com/ONSdigital/log.go/log"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

//go:generate moq -out testjob/job_queue.go -pkg testjob . Queue
//go:generate moq -out testjob/dataset_api.go -pkg testjob . DatasetAPI
//go:generate moq -out testjob/recipe_api.go -pkg testjob . RecipeAPI

// A list of custom errors
var (
	ErrGetRecipeFailed = errors.New("failed to get recipe")
	ErrSaveJobFailed   = errors.New("failed to save job")
)

// ErrCreateInstanceFailed builds the message for an error when creating an instance
func ErrCreateInstanceFailed(datasetID string) error {
	return fmt.Errorf("failed to create a new instance on the dataset api for: [%s]", datasetID)
}

// Service provides job related functionality.
type Service struct {
	dataStore        datastore.DataStorer
	queue            Queue
	datasetAPI       DatasetAPIClient
	recipeAPI        RecipeAPIClient
	urlBuilder       *url.Builder
	serviceAuthToken string
}

// Queue interface used to queue import jobs.
type Queue interface {
	Queue(ctx context.Context, job *models.ImportData) error
}

// DatasetAPIClient interface to the dataset API.
type DatasetAPIClient interface {
	// CreateInstance(ctx context.Context, job *models.Job, recipeInst *recipe.Instance) (instance *models.Instance, err error)
	PostInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instance *dataset.NewInstance) error
	//UpdateInstanceState(ctx context.Context, instanceID string, newState string) error
	PutInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string, instanceUpdate dataset.UpdateInstance) error
}

// RecipeAPIClient interface to the recipe API.
type RecipeAPIClient interface {
	GetRecipe(ctx context.Context, userAuthToken, serviceAuthToken, recipeID string) (*recipe.Recipe, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// NewService returns a new instance of a job.Service using the given dependencies.
func NewService(dataStore datastore.DataStorer, queue Queue, datasetAPI DatasetAPIClient, recipeAPI RecipeAPIClient, urlBuilder *url.Builder, serviceAuthToken string) *Service {
	return &Service{
		dataStore:        dataStore,
		queue:            queue,
		datasetAPI:       datasetAPI,
		recipeAPI:        recipeAPI,
		urlBuilder:       urlBuilder,
		serviceAuthToken: serviceAuthToken,
	}
}

// CreateJob creates a new job using the given job instance.
func (service Service) CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	logData := log.Data{"job": job}

	if err := job.Validate(); err != nil {
		log.Event(ctx, "CreateJob: failed validation", log.ERROR, log.Error(err), logData)
		return nil, errs.ErrInvalidJob
	}

	//Get details needed for instances from Recipe API
	recipe, err := service.recipeAPI.GetRecipe(ctx, "", "", job.RecipeID)
	if err != nil {
		log.Event(ctx, "CreateJob: failed to get recipe details", log.ERROR, log.Error(err), logData)
		return nil, ErrGetRecipeFailed
	}

	jobID, err := uuid.NewV4()
	if err != nil {
		log.Event(ctx, "CreateJob: failed to get UUID", log.ERROR, log.Error(err), logData)
		return nil, err
	}
	job.ID = jobID.String()
	job.Links.Self.HRef = service.urlBuilder.GetJobURL(job.ID)

	for _, oi := range recipe.OutputInstances {
		// now create an instance for this file
		// instance, instanceErr := service.datasetAPI.CreateInstance(ctx, job, &oi)

		path := api.URL + "/instances"
		datasetPath := api.URL + "/datasets/" + recipeInst.DatasetID
		logData := log.Data{"URL": path, "job_id": job.ID, "job_url": job.Links.Self.HRef}

		newInstance := models.CreateInstance(job, oi.DatasetID, datasetPath, oi.CodeLists)
		instance, err := service.datasetAPI.PostInstance(ctx, "", service.serviceAuthToken, "", newInstance)

		if err != nil {
			logData["instance"] = oi
			log.Event(ctx, "CreateJob: failed to create instance in datastore", log.ERROR, log.Error(err), logData)
			return nil, ErrCreateInstanceFailed(oi.DatasetID)
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
		log.Event(ctx, "CreateJob: failed to create job in datastore", log.ERROR, log.Error(err), logData)
		return nil, ErrSaveJobFailed
	}

	return createdJob, nil
}

// UpdateJob updates the job for the given jobID with the values in the given job model.
func (service Service) UpdateJob(ctx context.Context, jobID string, job *models.Job) error {

	err := service.dataStore.UpdateJob(jobID, job)
	if err != nil {
		return err
	}

	log.Event(ctx, "job updated", log.INFO, log.Data{"job": job, "job_id": jobID})
	if job.State == "submitted" {
		tasks, err := service.prepareJob(ctx, jobID)
		if err != nil {
			log.Event(ctx, "error preparing job", log.ERROR, log.Error(err), log.Data{"jobState": job, "job_id": jobID})
			return err
		}

		err = service.queue.Queue(ctx, tasks)
		if err != nil {
			log.Event(ctx, "error queueing tasks", log.ERROR, log.Error(err), log.Data{"tasks": tasks})
			return err
		}

		log.Event(ctx, "import job was queued", log.INFO, log.Data{"job": job, "job_id": jobID})
	}

	return nil
}

// PrepareJob returns a format ready to send to downstream services via kafka
func (service Service) prepareJob(ctx context.Context, jobID string) (*models.ImportData, error) {

	importJob, err := service.dataStore.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	recipe, err := service.recipeAPI.GetRecipe(ctx, "", "", importJob.RecipeID)
	if err != nil {
		return nil, err
	}

	var instanceIds []string
	for _, instanceRef := range importJob.Links.Instances {
		instanceIds = append(instanceIds, instanceRef.ID)

		err = service.datasetAPI.PutInstance(ctx, "", service.serviceAuthToken, "", instanceRef.ID,
			dataset.UpdateInstance{
				State: dataset.StateSubmitted.String(),
			},
		)
		if err != nil {
			// if err = service.datasetAPI.UpdateInstanceState(ctx, instanceRef.ID, "submitted"); err != nil {
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
