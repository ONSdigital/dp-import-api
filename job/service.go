package job

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/url"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

//go:generate moq -out testjob/job_queue.go -pkg testjob . Queue
//go:generate moq -out testjob/dataset_api.go -pkg testjob . DatasetAPIClient
//go:generate moq -out testjob/recipe_api.go -pkg testjob . RecipeAPIClient

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
	datasetAPIURL    string
	datasetAPIClient DatasetAPIClient
	recipeAPIClient  RecipeAPIClient
	urlBuilder       *url.Builder
	serviceAuthToken string
}

// Queue interface used to queue import jobs.
type Queue interface {
	Queue(ctx context.Context, job *models.ImportData) error
}

// DatasetAPIClient interface to the dataset API.
type DatasetAPIClient interface {
	PostInstance(ctx context.Context, serviceAuthToken string, newInstance *dataset.NewInstance) (instance *dataset.Instance, eTag string, err error)
	PutInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string, instance dataset.UpdateInstance, ifMatch string) (eTag string, err error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// RecipeAPIClient interface to the recipe API.
type RecipeAPIClient interface {
	GetRecipe(ctx context.Context, userAuthToken, serviceAuthToken, recipeID string) (*recipe.Recipe, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// NewService returns a new instance of a job.Service using the given dependencies.
func NewService(dataStore datastore.DataStorer, queue Queue, datasetAPIURL string, datasetAPIClient DatasetAPIClient, recipeAPIClient RecipeAPIClient, urlBuilder *url.Builder, serviceAuthToken string) *Service {
	return &Service{
		dataStore:        dataStore,
		queue:            queue,
		datasetAPIURL:    datasetAPIURL,
		datasetAPIClient: datasetAPIClient,
		recipeAPIClient:  recipeAPIClient,
		urlBuilder:       urlBuilder,
		serviceAuthToken: serviceAuthToken,
	}
}

// CreateJob creates a new job using the instances corresponding to the recipe defined by recipeID in the provided job.
// A new instance will be posted to dataset api for each outputInstance defined in the recipe.
// Note that the provided job will be modified (ID, links and counts will be updated).
func (service Service) CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	logData := log.Data{"job": job}

	// Validate job
	if err := job.Validate(); err != nil {
		log.Error(ctx, "CreateJob: failed validation", err, logData)
		return nil, errs.ErrInvalidJob
	}

	// Get details needed for instances from Recipe API
	jobRecipe, err := service.recipeAPIClient.GetRecipe(ctx, "", "", job.RecipeID)
	if err != nil {
		log.Error(ctx, "CreateJob: failed to get recipe details", err, logData)
		return nil, ErrGetRecipeFailed
	}

	// Generate a new random UUID
	jobID, err := uuid.NewV4()
	if err != nil {
		log.Error(ctx, "CreateJob: failed to get UUID", err, logData)
		return nil, err
	}

	// Update job ID and self link
	job.ID = jobID.String()
	job.Links.Self = models.IDLink{
		HRef: service.urlBuilder.GetJobURL(job.ID),
		ID:   job.ID,
	}

	job.Processed = []models.ProcessedInstances{}

	for _, oi := range jobRecipe.OutputInstances {

		// Create a new instance by sending a 'POST /instances' to dataset API
		datasetPath := service.datasetAPIURL + "/datasets/" + oi.DatasetID
		newInstance := models.CreateInstance(job, oi.DatasetID, datasetPath, oi.CodeLists)
		newInstance.Type = jobRecipe.Format
		instance, _, err := service.datasetAPIClient.PostInstance(ctx, service.serviceAuthToken, newInstance)
		if err != nil {
			log.Error(ctx, "CreateJob: failed to create instance in datastore", err, log.Data{"job_id": job.ID, "job_url": job.Links.Self.HRef, "instance": oi})
			return nil, ErrCreateInstanceFailed(oi.DatasetID)
		}

		// Append the new instance link to provided job
		job.Links.Instances = append(job.Links.Instances,
			models.IDLink{
				ID:   instance.ID,
				HRef: service.urlBuilder.GetInstanceURL(instance.ID),
			},
		)

		// Initialise the processed instances count for the current instance
		job.Processed = append(job.Processed,
			models.ProcessedInstances{
				ID:             instance.ID,
				RequiredCount:  len(oi.CodeLists),
				ProcessedCount: 0,
			},
		)
	}

	// Add job to dataStore
	createdJob, err := service.dataStore.AddJob(ctx, job)
	if err != nil {
		log.Error(ctx, "CreateJob: failed to create job in datastore", err, logData)
		return nil, ErrSaveJobFailed
	}

	return createdJob, nil
}

// UpdateJob updates the job for the given jobID with the values in the given job model.
func (service Service) UpdateJob(ctx context.Context, jobID string, job *models.Job) error {

	err := service.dataStore.UpdateJob(ctx, jobID, job)
	if err != nil {
		return err
	}

	log.Info(ctx, "job updated", log.Data{"job": job, "job_id": jobID})
	if job.State == "submitted" {
		tasks, err := service.prepareJob(ctx, jobID)
		if err != nil {
			log.Error(ctx, "error preparing job", err, log.Data{"jobState": job, "job_id": jobID})
			return err
		}

		err = service.queue.Queue(ctx, tasks)
		if err != nil {
			log.Error(ctx, "error queueing tasks", err, log.Data{"tasks": tasks})
			return err
		}

		log.Info(ctx, "import job was queued", log.Data{"job": job, "job_id": jobID})
	}

	return nil
}

// PrepareJob returns a format ready to send to downstream services via kafka
func (service Service) prepareJob(ctx context.Context, jobID string) (*models.ImportData, error) {

	importJob, err := service.dataStore.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	jobRecipe, err := service.recipeAPIClient.GetRecipe(ctx, "", "", importJob.RecipeID)
	if err != nil {
		return nil, err
	}

	var instanceIds []string
	for _, instanceRef := range importJob.Links.Instances {
		instanceIds = append(instanceIds, instanceRef.ID)

		_, err = service.datasetAPIClient.PutInstance(ctx, "", service.serviceAuthToken, "", instanceRef.ID,
			dataset.UpdateInstance{
				State: dataset.StateSubmitted.String(),
			},
			headers.IfMatchAnyETag,
		)
		if err != nil {
			return nil, err
		}
	}

	return &models.ImportData{
		JobID:         jobID,
		Recipe:        importJob.RecipeID,
		Format:        jobRecipe.Format,
		UploadedFiles: importJob.UploadedFiles,
		InstanceIDs:   instanceIds,
	}, nil
}
