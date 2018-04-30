package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/api/testapi"
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import-api/models"
	mockdatastore "github.com/ONSdigital/dp-import-api/mongo/testmongo"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/pkg/errors"
)

var (
	dstore              = mockdatastore.DataStorer{}
	dstoreNotFound      = mockdatastore.DataStorer{NotFound: true}
	dstoreInternalError = mockdatastore.DataStorer{InternalError: true}
	dummyJob            = &models.Job{ID: "34534543543"}
)

func TestAddJobReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("When the job service fails to save a job, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"recipe":"test"}`)
		r, err := createRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockJobService := &testapi.JobServiceMock{
			CreateJobFunc: func(ctx context.Context, newJob *models.Job) (*models.Job, error) {
				return nil, job.ErrSaveJobFailed
			},
		}

		auditorMock := newAuditorMock()

		api := CreateImportAPI(mux.NewRouter(), &dstoreInternalError, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})
}

func TestGetJobsReturnsInternalError(t *testing.T) {
	t.Parallel()

	auditorMock := newAuditorMock()

	Convey("When a get jobs request has no available datastore, an internal error is returned", t, func() {
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstoreInternalError, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, getJobsAuditAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)

		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, getJobsAuditAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, notFound)
	})
}

func TestGetJobs(t *testing.T) {
	t.Parallel()

	Convey("When a get jobs request has a datastore, an ok status is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		mockJobService := &testapi.JobServiceMock{}
		w := httptest.NewRecorder()
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, getJobsAuditAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)

		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, getJobsAuditAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, actionSuccessful)
	})
	Convey("When a get jobs request has a no auth token a 404 is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithOutAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		mockJobService := &testapi.JobServiceMock{}
		w := httptest.NewRecorder()
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("When a get jobs request is successful but auditing action attempted fails an internal server error status is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		mockJobService := &testapi.JobServiceMock{}
		w := httptest.NewRecorder()

		auditorMock = &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if result == actionAttempted {
					return errors.New("BOOM")
				}
				return nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(len(auditorMock.RecordCalls()), ShouldEqual, 1)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, getJobsAuditAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)
	})

	Convey("When a get jobs request is successful but auditing action successful fails an internal server error status is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		mockJobService := &testapi.JobServiceMock{}
		w := httptest.NewRecorder()

		auditorMock = &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if result == actionSuccessful {
					return errors.New("BOOM")
				}
				return nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, getJobsAuditAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)
		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, getJobsAuditAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, actionSuccessful)
	})
}

func TestGetJobReturnsNotFound(t *testing.T) {
	t.Parallel()

	Convey("When a get job request has a invalid jobID, a not found status is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs/000000", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstoreNotFound, mockJobService, auditorMock)

		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)

		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, getJobAuditAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)
		So(auditorMock.RecordCalls()[0].Params, ShouldResemble, common.Params{jobIDKey: "000000"})

		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, getJobAuditAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, notFound)
		So(auditorMock.RecordCalls()[1].Params, ShouldResemble, common.Params{jobIDKey: "000000"})
	})
}

func TestGetJob(t *testing.T) {
	t.Parallel()

	auditorMock := newAuditorMock()

	Convey("When a data store is available, an ok status is returned", t, func() {
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)

		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)

		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, getJobAuditAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)
		So(auditorMock.RecordCalls()[0].Params, ShouldResemble, common.Params{jobIDKey: "123"})

		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, getJobAuditAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, actionSuccessful)
		So(auditorMock.RecordCalls()[1].Params, ShouldResemble, common.Params{jobIDKey: "123"})
	})

	Convey("When no auth token is provided a 404 is returned", t, func() {
		r, err := createRequestWithOutAuth("GET", "http://localhost:21800/jobs/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestAddJobReturnsBadClientRequest(t *testing.T) {
	t.Parallel()

	auditorMock := newAuditorMock()

	Convey("When the job service returns an invalid job error, a bad request is returned", t, func() {
		reader := strings.NewReader("{ }")
		r, err := createRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockJobService := &testapi.JobServiceMock{
			CreateJobFunc: func(ctx context.Context, newJob *models.Job) (*models.Job, error) {
				return nil, job.ErrInvalidJob
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstoreNotFound, mockJobService, auditorMock)

		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)

		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, addJobAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)
		So(auditorMock.RecordCalls()[0].Params, ShouldResemble, common.Params{"recipeID": ""})

		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, addJobAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, actionUnsuccessful)
		So(auditorMock.RecordCalls()[1].Params, ShouldResemble, common.Params{"recipeID": ""})
	})
}

func TestAddJob(t *testing.T) {
	t.Parallel()

	Convey("When a valid message is sent, a jobInstance model is returned", t, func() {
		auditorMock := newAuditorMock()
		reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
		r, err := createRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		mockJobService = &testapi.JobServiceMock{
			CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
				return dummyJob, nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(w.Body.String(), ShouldContainSubstring, "\"id\":\"34534543543\"")

		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Params["recipeID"], ShouldEqual, "test")
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, addJobAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)

		So(auditorMock.RecordCalls()[1].Params["createdJobID"], ShouldEqual, dummyJob.ID)
		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, addJobAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, actionSuccessful)
	})

	Convey("when a job is created successfully but auditing action attempted returns an error then an error response is returned", t, func() {
		auditorMock := newAuditorMock()

		auditorMock.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
			return errors.New("auditing error")
		}

		reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
		r, err := createRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		mockJobService = &testapi.JobServiceMock{
			CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
				return dummyJob, nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(strings.TrimSpace(w.Body.String()), ShouldEqual, "Internal server error")

		So(len(auditorMock.RecordCalls()), ShouldEqual, 1)
		So(auditorMock.RecordCalls()[0].Params["recipeID"], ShouldEqual, "test")
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, addJobAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)
	})

	Convey("when a job is created successfully but auditing action successful then an error then an error response is returned", t, func() {
		auditorMock := newAuditorMock()
		auditorMock.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
			if result == actionSuccessful {
				return errors.New("auditing error")
			}
			return nil
		}

		reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
		r, err := createRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		mockJobService = &testapi.JobServiceMock{
			CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
				return dummyJob, nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(strings.TrimSpace(w.Body.String()), ShouldEqual, "Internal server error")

		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Params["recipeID"], ShouldEqual, "test")
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, addJobAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)

		So(auditorMock.RecordCalls()[1].Params["createdJobID"], ShouldEqual, dummyJob.ID)
		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, addJobAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, actionSuccessful)
	})
}

func TestAddS3FileReturnsNotFound(t *testing.T) {
	t.Parallel()

	auditorMock := newAuditorMock()

	Convey("When adding a S3 file to an import queue job with a invalid instance id, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstoreNotFound, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestUpdateJobState(t *testing.T) {
	t.Parallel()

	Convey("when updating a job state and audit action attempted fails an internal server error is returned", t, func() {
		auditorMock := newAuditorMock()
		auditorMock.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
			return errors.New("broken")
		}

		reader := strings.NewReader("{ \"state\":\"start\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockJobService := &testapi.JobServiceMock{
			UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
				return nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(auditorMock.RecordCalls()), ShouldEqual, 1)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, updateJobAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)
		So(auditorMock.RecordCalls()[0].Params, ShouldResemble, common.Params{jobIDKey: "12345"})
		So(len(mockJobService.UpdateJobCalls()), ShouldEqual, 0)
	})

	Convey("When successfully updating a jobs state, it returns an OK code", t, func() {
		auditorMock := newAuditorMock()
		reader := strings.NewReader("{ \"state\":\"start\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockJobService := &testapi.JobServiceMock{
			UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
				return nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, updateJobAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)
		So(auditorMock.RecordCalls()[0].Params, ShouldResemble, common.Params{jobIDKey: "12345"})
		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, updateJobAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, actionSuccessful)
		So(auditorMock.RecordCalls()[1].Params, ShouldResemble, common.Params{jobIDKey: "12345"})
	})

	Convey("When updating a jobs state with no auth token, it returns an not found code", t, func() {
		auditorMock := newAuditorMock()
		reader := strings.NewReader("{ \"state\":\"start\"}")
		r, err := createRequestWithOutAuth("PUT", "http://localhost:21800/jobs/12345", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockJobService := &testapi.JobServiceMock{
			UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
				return nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestUpdateJobStateReturnsNotFound(t *testing.T) {
	t.Parallel()

	auditorMock := newAuditorMock()

	Convey("When the job service returns a job not found error, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"state\":\"start\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		mockJobService = &testapi.JobServiceMock{
			UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
				return api_errors.JobNotFoundError
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstoreNotFound, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)

		So(len(auditorMock.RecordCalls()), ShouldEqual, 2)
		So(auditorMock.RecordCalls()[0].Action, ShouldEqual, updateJobAction)
		So(auditorMock.RecordCalls()[0].Result, ShouldEqual, actionAttempted)
		So(auditorMock.RecordCalls()[0].Params, ShouldResemble, common.Params{jobIDKey: "12345"})
		So(auditorMock.RecordCalls()[1].Action, ShouldEqual, updateJobAction)
		So(auditorMock.RecordCalls()[1].Result, ShouldEqual, notFound)
		So(auditorMock.RecordCalls()[1].Params, ShouldResemble, common.Params{jobIDKey: "12345"})
	})
}

func TestUpdateJobStateToSubmitted(t *testing.T) {
	t.Parallel()

	auditorMock := newAuditorMock()

	Convey("When a job state is updated to submitted, a message is sent into the job queue", t, func() {
		reader := strings.NewReader("{ \"state\":\"submitted\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockJobService := &testapi.JobServiceMock{
			UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
				return nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func createRequestWithAuth(method, URL string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, URL, body)
	ctx := r.Context()
	ctx = identity.SetCaller(ctx, "someone@ons.gov.uk")
	r = r.WithContext(ctx)
	return r, err
}

func createRequestWithOutAuth(method, URL string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, URL, body)
	return r, err
}

func newAuditorMock() *audit.AuditorServiceMock {
	return &audit.AuditorServiceMock{
		RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
			log.Debug("capturing audit event", nil)
			return nil
		},
	}
}
