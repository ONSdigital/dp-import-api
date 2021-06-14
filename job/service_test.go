package job_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/job/testjob"
	"github.com/ONSdigital/dp-import-api/models"
	mongo "github.com/ONSdigital/dp-import-api/mongo/testmongo"
	"github.com/ONSdigital/dp-import-api/url"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	urlBuilder  = url.NewBuilder("http://import-api", "http://dataset-api")
	dummyRecipe = &recipe.Recipe{
		ID:    "123",
		Alias: "alias",
		OutputInstances: []recipe.Instance{
			{
				DatasetID: "dataset1",
			},
			{
				DatasetID: "dataset2",
			},
		},
	}
	dummyInstance = &dataset.Instance{
		Version: dataset.Version{
			InstanceID: "654",
		},
	}
	datasetAPIURL    = "http://localhost:22000"
	serviceAuthToken = "testToken"
	ctx              = context.Background()
)

func TestService_CreateJob(t *testing.T) {

	Convey("Given a job service with mocked dependencies", t, func() {

		mockDataStore := &mongo.DataStorer{}
		mockedQueue := &testjob.QueueMock{}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{
			PostInstanceFunc: func(ctx context.Context, serviceAuthToken string, newInstance *dataset.NewInstance) (*dataset.Instance, error) {
				return dummyInstance, nil
			},
			PutInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, instanceUpdate dataset.UpdateInstance) error {
				return nil
			},
		}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{
			GetRecipeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error) {
				return dummyRecipe, nil
			},
		}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		job := &models.Job{
			RecipeID: "123-234-456",
		}

		Convey("When create job is called", func() {

			createdJob, err := jobService.CreateJob(ctx, job)

			Convey("The expected calls are made to dependencies and the job is updated", func() {
				So(err, ShouldBeNil)
				So(createdJob, ShouldNotBeNil)

				So(job.ID, ShouldNotBeBlank)
				So(job.Links.Self.HRef, ShouldNotBeBlank)

				// an instance is created for each output instance specified in the recipe.
				So(len(mockedDatasetAPI.PostInstanceCalls()), ShouldEqual, len(dummyRecipe.OutputInstances))
				So(len(job.Links.Instances), ShouldEqual, len(dummyRecipe.OutputInstances))
			})
		})
	})
}

func TestService_CreateJob_CreateInstanceFails(t *testing.T) {

	Convey("Given a job service with a mock dataset API that returns a failure", t, func() {

		mockDataStore := &mongo.DataStorer{}
		mockedQueue := &testjob.QueueMock{}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{
			PostInstanceFunc: func(ctx context.Context, serviceAuthToken string, newInstance *dataset.NewInstance) (*dataset.Instance, error) {
				return nil, errors.New("Create instance failed.")
			},
		}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{
			GetRecipeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error) {
				return dummyRecipe, nil
			},
		}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		newJob := &models.Job{
			RecipeID: "123-234-456",
		}

		Convey("When create job is called", func() {

			createdJob, err := jobService.CreateJob(ctx, newJob)

			Convey("The expected error is returned", func() {
				So(err, ShouldResemble, job.ErrCreateInstanceFailed("dataset1"))
				So(createdJob, ShouldBeNil)
			})
		})
	})
}

func TestService_CreateJob_SaveJobFails(t *testing.T) {

	Convey("Given a job service with mocked dependencies", t, func() {

		mockDataStore := &mongo.DataStorer{InternalError: true}
		mockedQueue := &testjob.QueueMock{}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{
			PostInstanceFunc: func(ctx context.Context, serviceAuthToken string, newInstance *dataset.NewInstance) (*dataset.Instance, error) {
				return dummyInstance, nil
			},
		}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{
			GetRecipeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error) {
				return dummyRecipe, nil
			},
		}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		newJob := &models.Job{
			RecipeID: "123-234-456",
		}

		Convey("When create job is called", func() {

			createdJob, err := jobService.CreateJob(ctx, newJob)

			Convey("The expected error is returned", func() {
				So(err, ShouldEqual, job.ErrSaveJobFailed)
				So(createdJob, ShouldBeNil)
			})
		})
	})
}

func TestService_CreateJob_InvalidJob(t *testing.T) {

	Convey("Given a job service with mocked dependencies", t, func() {

		mockDataStore := &mongo.DataStorer{}
		mockedQueue := &testjob.QueueMock{}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		Convey("When a job with no recipe URL is passed to create job", func() {

			newJob := &models.Job{}

			createdJob, err := jobService.CreateJob(ctx, newJob)

			Convey("Then an invalid job error is returned ", func() {
				So(err, ShouldEqual, errs.ErrInvalidJob)
				So(createdJob, ShouldBeNil)
			})
		})
	})
}

func TestService_CreateJob_GetRecipeFails(t *testing.T) {

	Convey("Given a job service with a mock recipe API that returns an error", t, func() {

		mockDataStore := &mongo.DataStorer{}
		mockedQueue := &testjob.QueueMock{}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{
			GetRecipeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error) {
				return nil, errors.New("the server melted")
			},
		}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		newJob := &models.Job{
			RecipeID: "123-234-456",
		}

		Convey("When create job is called", func() {

			createdJob, err := jobService.CreateJob(ctx, newJob)

			Convey("Then the service returns the expected error ", func() {
				So(err, ShouldEqual, job.ErrGetRecipeFailed)
				So(createdJob, ShouldBeNil)
			})
		})
	})
}

func TestService_UpdateJob(t *testing.T) {

	Convey("Given a job service with mocked dependencies", t, func() {

		mockDataStore := &mongo.DataStorer{}
		mockedQueue := &testjob.QueueMock{
			QueueFunc: func(ctx context.Context, job *models.ImportData) error {
				return nil
			},
		}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{
			PutInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, instanceUpdate dataset.UpdateInstance) error {
				return nil
			},
		}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		jobID := "123"
		job := &models.Job{
			RecipeID: "123-234-456",
			ID:       jobID,
		}

		Convey("When update job is called", func() {

			err := jobService.UpdateJob(ctx, jobID, job)

			Convey("The expected calls are made to dependencies", func() {
				So(err, ShouldBeNil)
				So(len(mockedQueue.QueueCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestService_UpdateJob_SaveFails(t *testing.T) {

	Convey("Given a job service with mocked dependencies", t, func() {

		mockDataStore := &mongo.DataStorer{InternalError: true}
		mockedQueue := &testjob.QueueMock{
			QueueFunc: func(ctx context.Context, job *models.ImportData) error {
				return nil
			},
		}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{
			PutInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, instanceUpdate dataset.UpdateInstance) error {
				return nil
			},
		}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		jobID := "123"
		updatedJob := &models.Job{
			RecipeID: "123-234-456",
			ID:       jobID,
		}

		Convey("When update job is called", func() {

			err := jobService.UpdateJob(ctx, jobID, updatedJob)
			Convey("The expected calls are made to dependencies", func() {
				So(err, ShouldEqual, mongo.InternalError)
				So(len(mockedQueue.QueueCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestService_UpdateJob_QueuesWhenSubmitted(t *testing.T) {

	Convey("Given a job service with mocked dependencies", t, func() {

		mockDataStore := &mongo.DataStorer{}
		mockedQueue := &testjob.QueueMock{
			QueueFunc: func(ctx context.Context, job *models.ImportData) error {
				return nil
			},
		}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{
			PutInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, instanceUpdate dataset.UpdateInstance) error {
				return nil
			},
		}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{
			GetRecipeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error) {
				return dummyRecipe, nil
			},
		}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		jobID := "123"
		job := &models.Job{
			RecipeID: "123-234-456",
			ID:       jobID,
			State:    "submitted",
		}

		Convey("When update job is called", func() {

			err := jobService.UpdateJob(ctx, jobID, job)

			Convey("The expected calls are made to dependencies", func() {
				So(err, ShouldBeNil)
				So(len(mockedRecipeAPI.GetRecipeCalls()), ShouldEqual, 1)
				So(len(mockedQueue.QueueCalls()), ShouldEqual, 1)
			})
		})
	})
}
