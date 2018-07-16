package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-import-api/api/testapi"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"io"
)

var dummyJob = &models.Job{ID: "34534543543"}

func TestFailureToAddJob(t *testing.T) {

	t.Parallel()
	attemptedAuditParams := common.Params{"caller_identity": "someone@ons.gov.uk"}

	Convey("Given a request to add a job", t, func() {
		Convey("When no auth token is provided", func() {
			Convey("Then return status unauthorised (401)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
				r, err := testapi.CreateRequestWithOutAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnauthorised.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], addJobAction, audit.Attempted, common.Params{})
				testapi.VerifyAuditorCalls(calls[1], addJobAction, audit.Unsuccessful, common.Params{})

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey("When the request body is invalid", func() {
			Convey("Then return status bad request (400)", func() {
				mockJobService := &testapi.JobServiceMock{}
				auditorMock := testapi.NewAuditorMock()
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)

				reader := strings.NewReader("{")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrFailedToParseJSONBody.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], addJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], addJobAction, audit.Unsuccessful, nil)

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey("When the request body is missing a mandatory field", func() {
			Convey("Then return status bad request (400)", func() {
				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, newJob *models.Job) (*models.Job, error) {
						return nil, errs.ErrInvalidJob
					},
				}
				auditorMock := testapi.NewAuditorMock()
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"number_of_instances\": 1}")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInvalidJob.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], addJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], addJobAction, audit.Unsuccessful, common.Params{"recipeID": ""})

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey("When the import api is unable to connect to its datastore", func() {
			Convey("Then return status internal server error (500)", func() {
				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, newJob *models.Job) (*models.Job, error) {
						return nil, errs.ErrInternalServer
					},
				}
				auditorMock := testapi.NewAuditorMock()
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreInternalError, mockJobService, auditorMock)

				reader := strings.NewReader(`{"recipe":"test"}`)
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], addJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], addJobAction, audit.Unsuccessful, common.Params{"recipeID": "test"})

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey("When auditing request for the attempted action fails", func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						return errors.New("auditing error")
					},
				}
				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
						return dummyJob, nil
					},
				}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 1)

				So(len(auditorMock.RecordCalls()), ShouldEqual, 1)
				testapi.VerifyAuditorCalls(calls[0], addJobAction, audit.Attempted, attemptedAuditParams)

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey(`When request body is missing a mandatory field but the
			auditing of unsuccessful action fails`, func() {

			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Unsuccessful {
							return errors.New("auditing error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}

				reader := strings.NewReader("{ ")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)
				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], addJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], addJobAction, audit.Unsuccessful, nil)

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey(`When creating the new job a duplication key error occurs but the
			auditing of unsuccessful action fails`, func() {

			Convey("Then return status internal server error (500)", func() {
				auditorMock := testapi.NewAuditorMock()
				auditorMock.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
					if result == audit.Unsuccessful {
						return errors.New("auditing error")
					}
					return nil
				}

				reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
						return nil, errors.New("duplication key error")
					},
				}

				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)
				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], addJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], addJobAction, audit.Unsuccessful, common.Params{"recipeID": "test"})

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})
	})
}

func TestSuccessfullyAddJob(t *testing.T) {

	t.Parallel()
	attemptedAuditParams := common.Params{"caller_identity": "someone@ons.gov.uk"}

	Convey("Given a valid request to add a job", t, func() {
		Convey("When successfully created in datastore", func() {
			Convey("Then return status created (201)", func() {
				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
						return dummyJob, nil
					},
				}
				auditorMock := testapi.NewAuditorMock()
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusCreated)
				So(w.Body.String(), ShouldContainSubstring, "\"id\":\"34534543543\"")

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				p := common.Params{"createdJobID": dummyJob.ID, "recipeID": "test"}
				testapi.VerifyAuditorCalls(calls[0], addJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], addJobAction, audit.Successful, p)

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey("When auditing request but the successful action fails", func() {
			Convey("Then continue to return status created (201)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Successful {
							return errors.New("auditing error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
						return dummyJob, nil
					},
				}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusCreated)
				So(w.Body.String(), ShouldContainSubstring, "\"id\":\"34534543543\"")

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], addJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], addJobAction, audit.Successful, common.Params{"recipeID": "test", "createdJobID": dummyJob.ID})

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})
	})
}
