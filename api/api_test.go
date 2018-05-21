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

func verifyAuditorCalls(callInfo struct {
	Ctx    context.Context
	Action string
	Result string
	Params common.Params
}, a string, r string, p common.Params) {
	So(callInfo.Action, ShouldEqual, a)
	So(callInfo.Result, ShouldEqual, r)
	So(callInfo.Params, ShouldResemble, p)
}

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

		calls := auditorMock.RecordCalls()

		So(len(calls), ShouldEqual, 2)
		p := common.Params{}
		verifyAuditorCalls(calls[0], getJobsAction, actionAttempted, p)
		verifyAuditorCalls(calls[1], getJobsAction, actionUnsuccessful, p)
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

		calls := auditorMock.RecordCalls()
		p := common.Params{}
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], getJobsAction, actionAttempted, p)
		verifyAuditorCalls(calls[1], getJobsAction, actionSuccessful, p)
	})

	Convey("When a get jobs request has a no auth token a 401 is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithOutAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		mockJobService := &testapi.JobServiceMock{}
		w := httptest.NewRecorder()
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
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

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 1)
		verifyAuditorCalls(calls[0], getJobsAction, actionAttempted, common.Params{})
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

		calls := auditorMock.RecordCalls()
		p := common.Params{}
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], getJobsAction, actionAttempted, p)
		verifyAuditorCalls(calls[1], getJobsAction, actionSuccessful, p)
	})

	Convey("When a datastore.getJobs returns an error then a 500 status is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		mockJobService := &testapi.JobServiceMock{}
		w := httptest.NewRecorder()

		auditorMock = &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				return nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &mockdatastore.DataStorer{InternalError: true}, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		calls := auditorMock.RecordCalls()
		p := common.Params{}
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], getJobsAction, actionAttempted, p)
		verifyAuditorCalls(calls[1], getJobsAction, actionUnsuccessful, p)
	})

	Convey("When a datastore.getJobs returns an error and auditing action unsuccessful errors then a 500 status is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		mockJobService := &testapi.JobServiceMock{}
		w := httptest.NewRecorder()

		auditorMock = &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getJobsAction && result == actionUnsuccessful {
					return errors.New("audit error")
				}
				return nil
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &mockdatastore.DataStorer{InternalError: true}, mockJobService, auditorMock)

		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		calls := auditorMock.RecordCalls()
		p := common.Params{}
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], getJobsAction, actionAttempted, p)
		verifyAuditorCalls(calls[1], getJobsAction, actionUnsuccessful, p)
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

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], getJobAction, actionAttempted, common.Params{jobIDKey: "000000"})
		verifyAuditorCalls(calls[1], getJobAction, actionUnsuccessful, common.Params{jobIDKey: "000000"})
	})
}

func TestGetJob(t *testing.T) {
	t.Parallel()

	Convey("When a data store is available, an ok status is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)

		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], getJobAction, actionAttempted, common.Params{jobIDKey: "123"})
		verifyAuditorCalls(calls[1], getJobAction, actionSuccessful, common.Params{jobIDKey: "123"})
	})

	Convey("When no auth token is provided a 401 is returned", t, func() {
		auditorMock := newAuditorMock()
		r, err := createRequestWithOutAuth("GET", "http://localhost:21800/jobs/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})

	Convey("When get job audit action attempted errors then a 500 status is returned", t, func() {
		auditorMock := &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getJobAction && result == actionAttempted {
					return errors.New("audit error")
				}
				return nil
			},
		}

		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)

		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 1)
		verifyAuditorCalls(calls[0], getJobAction, actionAttempted, common.Params{jobIDKey: "123"})
	})

	Convey("When datastore.getJob returns an error and audit action unsuccessful errors then a 500 status is returned", t, func() {
		auditorMock := &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getJobAction && result == actionUnsuccessful {
					return errors.New("audit error")
				}
				return nil
			},
		}

		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstoreInternalError, mockJobService, auditorMock)

		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], getJobAction, actionAttempted, common.Params{jobIDKey: "123"})
		verifyAuditorCalls(calls[1], getJobAction, actionUnsuccessful, common.Params{jobIDKey: "123"})
	})

	Convey("When get job is successful but auditing action successful errors then a 500 status is returned", t, func() {
		auditorMock := &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getJobAction && result == actionSuccessful {
					return errors.New("audit error")
				}
				return nil
			},
		}

		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)

		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], getJobAction, actionAttempted, common.Params{jobIDKey: "123"})
		verifyAuditorCalls(calls[1], getJobAction, actionSuccessful, common.Params{jobIDKey: "123"})
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

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], addJobAction, actionAttempted, common.Params{"recipeID": ""})
		verifyAuditorCalls(calls[1], addJobAction, actionUnsuccessful, common.Params{"recipeID": ""})
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

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], addJobAction, actionAttempted, common.Params{"createdJobID": dummyJob.ID, "recipeID": "test"})
		verifyAuditorCalls(calls[1], addJobAction, actionSuccessful, common.Params{"createdJobID": dummyJob.ID, "recipeID": "test"})
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
		So(strings.TrimSpace(w.Body.String()), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 1)

		So(len(auditorMock.RecordCalls()), ShouldEqual, 1)
		verifyAuditorCalls(calls[0], addJobAction, actionAttempted, common.Params{"recipeID": "test"})
	})

	Convey("when a job is created successfully but auditing action successful errors then an error response is returned", t, func() {
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
		So(strings.TrimSpace(w.Body.String()), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], addJobAction, actionAttempted, common.Params{"recipeID": "test", "createdJobID": dummyJob.ID})
		verifyAuditorCalls(calls[1], addJobAction, actionSuccessful, common.Params{"recipeID": "test", "createdJobID": dummyJob.ID})
	})

	Convey("when jobservice.createjob returns an error and auditing action unssucessful errors then a 500 status is returned", t, func() {
		auditorMock := newAuditorMock()
		auditorMock.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
			if result == actionUnsuccessful {
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
				return nil, errors.New("create job error")
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(strings.TrimSpace(w.Body.String()), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], addJobAction, actionAttempted, common.Params{"recipeID": "test"})
		verifyAuditorCalls(calls[1], addJobAction, actionUnsuccessful, common.Params{"recipeID": "test"})
	})
}

func TestAddFile(t *testing.T) {
	t.Parallel()

	params := common.Params{
		jobIDKey:    "12345",
		"fileAlias": "n1",
		"fileURL":   "https://aws.s3/ons/myfile.exel",
	}

	Convey("When adding a S3 file to an import queue job with a invalid instance id, it returns a not found code", t, func() {
		auditorMock := newAuditorMock()
		reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstoreNotFound, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)

		verifyAuditorCalls(calls[0], uploadFileAction, actionAttempted, params)
		verifyAuditorCalls(calls[1], uploadFileAction, actionUnsuccessful, params)
	})

	Convey("when auditing action attempted errors then a 500 status is returned", t, func() {
		auditorMock := &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if result == actionAttempted {
					return errors.New("audit error")
				}
				return nil
			},
		}
		reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstoreNotFound, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 1)

		verifyAuditorCalls(calls[0], uploadFileAction, actionAttempted, common.Params{jobIDKey: "12345"})
	})

	Convey("when dataStore.AddUploadedFile errors and auditing action unsuccessful errors then a 500 status is returned", t, func() {
		auditorMock := &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if result == actionUnsuccessful {
					return errors.New("audit error")
				}
				return nil
			},
		}
		reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstoreNotFound, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], uploadFileAction, actionAttempted, params)
		verifyAuditorCalls(calls[1], uploadFileAction, actionUnsuccessful, params)
	})

	Convey("when upload is successful but auditing action successful errors then a 500 status is returned", t, func() {
		auditorMock := &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if result == actionSuccessful {
					return errors.New("audit error")
				}
				return nil
			},
		}
		reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], uploadFileAction, actionAttempted, params)
		verifyAuditorCalls(calls[1], uploadFileAction, actionSuccessful, params)
	})

	Convey("when upload is successful then a 200 status is returned", t, func() {
		auditorMock := &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				return nil
			},
		}
		reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockJobService := &testapi.JobServiceMock{}
		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], uploadFileAction, actionAttempted, params)
		verifyAuditorCalls(calls[1], uploadFileAction, actionSuccessful, params)
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
		So(w.Body.String(), ShouldEqual, internalError)

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 1)
		verifyAuditorCalls(calls[0], updateJobAction, actionAttempted, common.Params{jobIDKey: "12345"})
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

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], updateJobAction, actionAttempted, common.Params{jobIDKey: "12345"})
		verifyAuditorCalls(calls[1], updateJobAction, actionSuccessful, common.Params{jobIDKey: "12345"})
	})

	Convey("When updating a jobs state with no auth token, it returns a 401", t, func() {
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
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
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

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], updateJobAction, actionAttempted, common.Params{jobIDKey: "12345"})
		verifyAuditorCalls(calls[1], updateJobAction, actionUnsuccessful, common.Params{jobIDKey: "12345"})
	})
}

func TestUpdateJobStateToSubmitted(t *testing.T) {
	t.Parallel()

	Convey("When a job state is updated to submitted, a message is sent into the job queue", t, func() {
		auditorMock := newAuditorMock()
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

		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], updateJobAction, actionAttempted, common.Params{jobIDKey: "12345"})
		verifyAuditorCalls(calls[1], updateJobAction, actionSuccessful, common.Params{jobIDKey: "12345"})
	})

	Convey("When a jobService.UpdateJob errors and auditing action unsuccessful returns an error then a 500 status is returned", t, func() {
		auditorMock := &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if result == actionUnsuccessful {
					return errors.New("audit error")
				}
				return nil
			},
		}
		reader := strings.NewReader("{ \"state\":\"submitted\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockJobService := &testapi.JobServiceMock{
			UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
				return errors.New("update job error")
			},
		}

		api := CreateImportAPI(mux.NewRouter(), &dstore, mockJobService, auditorMock)
		api.router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldEqual, internalError)
		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], updateJobAction, actionAttempted, common.Params{jobIDKey: "12345"})
		verifyAuditorCalls(calls[1], updateJobAction, actionUnsuccessful, common.Params{jobIDKey: "12345"})
	})

	Convey("When a job state is updated to submitted but auditing action successful returns an error then a 500 status is returned", t, func() {
		auditorMock := &audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if result == actionSuccessful {
					return errors.New("audit error")
				}
				return nil
			},
		}
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

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldEqual, internalError)
		calls := auditorMock.RecordCalls()
		So(len(calls), ShouldEqual, 2)
		verifyAuditorCalls(calls[0], updateJobAction, actionAttempted, common.Params{jobIDKey: "12345"})
		verifyAuditorCalls(calls[1], updateJobAction, actionSuccessful, common.Params{jobIDKey: "12345"})
	})
}

func createRequestWithAuth(method, URL string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, URL, body)
	ctx := r.Context()
	ctx = common.SetCaller(ctx, "someone@ons.gov.uk")
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
