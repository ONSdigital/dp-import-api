package job_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/job/testjob"
	"github.com/ONSdigital/dp-import-api/models"
	mongo "github.com/ONSdigital/dp-import-api/mongo/testmongo"
	"github.com/ONSdigital/dp-import-api/url"
	. "github.com/smartystreets/goconvey/convey"
)

const testETag = "testETag"

var (
	urlBuilder  = url.NewBuilder("http://import-api", "http://dataset-api")
	dummyRecipe = &recipe.Recipe{
		ID:    "123",
		Alias: "alias",
		OutputInstances: []recipe.Instance{
			{
				DatasetID:       "dataset1",
				CodeLists:       []recipe.CodeList{{ID: "codelist11"}, {ID: "codelist12"}},
				LowestGeography: "lowest_geo",
			},
			{
				DatasetID: "dataset2",
				CodeLists: []recipe.CodeList{{ID: "codelist21"}, {ID: "codelist22"}, {ID: "codelist23"}},
			},
		},
		Format: "cantabular_blob",
	}
	datasetAPIURL    = "http://localhost:22000"
	serviceAuthToken = "testToken"
	ctx              = context.Background()
)

// dummyInstance generates a dataset Instance for testing
func dummyInstance() *dataset.Instance {
	return &dataset.Instance{
		Version: dataset.Version{
			ID: "dummyInstanceID",
		},
	}
}

// expectedNewInstance creates an expected NewInstance for the provided jobID and datasetID
func expectedNewInstance(jobID, datasetID string) *dataset.NewInstance {
	newInstance := &dataset.NewInstance{
		Links: &dataset.Links{
			Dataset: dataset.Link{
				URL: "http://localhost:22000/datasets/" + datasetID,
				ID:  datasetID,
			},
			Job: dataset.Link{
				URL: "http://import-api/jobs/" + jobID,
				ID:  jobID,
			},
		},
		Dimensions: []dataset.CodeList{},
		ImportTasks: &dataset.InstanceImportTasks{
			ImportObservations: &dataset.ImportObservationsTask{
				State: dataset.StateCreated.String(),
			},
			BuildHierarchyTasks:   []*dataset.BuildHierarchyTask{},
			BuildSearchIndexTasks: []*dataset.BuildSearchIndexTask{},
		},
		Type: "cantabular_blob",
	}
	if datasetID == "dataset1" {
		newInstance.Dimensions = []dataset.CodeList{{ID: "codelist11"}, {ID: "codelist12"}}
		newInstance.LowestGeography = "lowest_geo"
	} else if datasetID == "dataset2" {
		newInstance.Dimensions = []dataset.CodeList{{ID: "codelist21"}, {ID: "codelist22"}, {ID: "codelist23"}}
	}
	return newInstance
}

func TestService_CreateJob(t *testing.T) {

	Convey("Given a job service with mocked dependencies", t, func() {

		mockDataStore := &mongo.DataStorer{}
		mockedQueue := &testjob.QueueMock{}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{
			PostInstanceFunc: func(ctx context.Context, serviceAuthToken string, newInstance *dataset.NewInstance) (*dataset.Instance, string, error) {
				retInstance := dummyInstance()
				retInstance.ID = "dummyInstance_" + newInstance.Links.Dataset.ID
				return retInstance, testETag, nil
			},
			PutInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, i dataset.UpdateInstance, ifMatch string) (string, error) {
				return testETag, nil
			},
		}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{
			GetRecipeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error) {
				return dummyRecipe, nil
			},
		}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		jobModel := &models.Job{
			RecipeID: "123-234-456",
		}

		Convey("When create job is called", func() {

			createdJob, err := jobService.CreateJob(ctx, jobModel)

			Convey("Then the expected job is returned without error", func() {
				So(createdJob, ShouldResemble, &mongo.CreatedJob)
				So(err, ShouldBeNil)
			})

			Convey("Then the provided job is mutated with the expected ID and link values", func() {
				So(jobModel.ID, ShouldNotBeBlank)
				So(jobModel.Links, ShouldResemble, &models.LinksMap{
					Self: models.IDLink{
						ID:   jobModel.ID,
						HRef: "http://import-api/jobs/" + jobModel.ID,
					},
					Instances: []models.IDLink{
						{
							ID:   "dummyInstance_dataset1",
							HRef: "http://dataset-api/instances/dummyInstance_dataset1",
						},
						{
							ID:   "dummyInstance_dataset2",
							HRef: "http://dataset-api/instances/dummyInstance_dataset2",
						},
					},
				})
			})

			Convey("Then the expected instances, as defined by the recipe, are posted to dataset API with the correct authentication", func() {
				So(mockedDatasetAPI.PostInstanceCalls(), ShouldHaveLength, 2)
				So(mockedDatasetAPI.PostInstanceCalls()[0].NewInstance, ShouldResemble, expectedNewInstance(jobModel.ID, "dataset1"))
				So(mockedDatasetAPI.PostInstanceCalls()[0].ServiceAuthToken, ShouldResemble, serviceAuthToken)
				So(mockedDatasetAPI.PostInstanceCalls()[1].NewInstance, ShouldResemble, expectedNewInstance(jobModel.ID, "dataset2"))
				So(mockedDatasetAPI.PostInstanceCalls()[1].ServiceAuthToken, ShouldResemble, serviceAuthToken)
			})
		})
	})
}

func TestService_CreateJob_CreateInstanceFails(t *testing.T) {

	Convey("Given a job service with a mock dataset API that returns a failure", t, func() {

		mockDataStore := &mongo.DataStorer{}
		mockedQueue := &testjob.QueueMock{}
		mockedDatasetAPI := &testjob.DatasetAPIClientMock{
			PostInstanceFunc: func(ctx context.Context, serviceAuthToken string, newInstance *dataset.NewInstance) (*dataset.Instance, string, error) {
				return nil, "", errors.New("Create instance failed.")
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
			PostInstanceFunc: func(ctx context.Context, serviceAuthToken string, newInstance *dataset.NewInstance) (*dataset.Instance, string, error) {
				return dummyInstance(), testETag, nil
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
			PutInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, i dataset.UpdateInstance, ifMatch string) (string, error) {
				return testETag, nil
			},
		}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		jobID := "123"
		jobUpdate := &models.Job{
			RecipeID: "123-234-456",
			ID:       jobID,
		}

		Convey("When update job is called", func() {

			err := jobService.UpdateJob(ctx, jobID, jobUpdate)

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
			PutInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, i dataset.UpdateInstance, ifMatch string) (string, error) {
				return testETag, nil
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
			PutInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, i dataset.UpdateInstance, ifMatch string) (string, error) {
				return testETag, nil
			},
		}
		mockedRecipeAPI := &testjob.RecipeAPIClientMock{
			GetRecipeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error) {
				return dummyRecipe, nil
			},
		}

		jobService := job.NewService(mockDataStore, mockedQueue, datasetAPIURL, mockedDatasetAPI, mockedRecipeAPI, urlBuilder, serviceAuthToken)

		jobID := "123"
		jobUpdate := &models.Job{
			RecipeID: "123-234-456",
			ID:       jobID,
			State:    "submitted",
		}

		Convey("When update job is called", func() {

			err := jobService.UpdateJob(ctx, jobID, jobUpdate)

			Convey("The expected calls are made to dependencies", func() {
				So(err, ShouldBeNil)
				So(len(mockedRecipeAPI.GetRecipeCalls()), ShouldEqual, 1)
				So(len(mockedQueue.QueueCalls()), ShouldEqual, 1)
			})
		})
	})
}
