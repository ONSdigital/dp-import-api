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
)

func TestFailureToUpdateJobState(t *testing.T) {
	t.Parallel()
	Convey("Given a request to update job state", t, func() {
		auditorMock := testapi.NewAuditorMock()

		Convey("When request has no auth header", func() {
			Convey("Then return status unauthorised (401)", func() {
				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return nil
					},
				}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{\"state\":\"created\"}")
				r, err := testapi.CreateRequestWithOutAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnauthorised.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 0)
			})
		})

		Convey("When request contains an invalid body", func() {
			Convey("Then return status bad request (400)", func() {
				reader := strings.NewReader("{")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return errs.ErrJobNotFound
					},
				}

				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrFailedToParseJSONBody.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], updateJobAction, audit.Attempted, common.Params{jobIDKey: "12345"})
				testapi.VerifyAuditorCalls(calls[1], updateJobAction, audit.Unsuccessful, common.Params{jobIDKey: "12345"})
			})
		})

		Convey("When request body contains an invalid state value", func() {
			Convey("Then return status bad request (400)", func() {
				reader := strings.NewReader("{\"state\":\"start\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return errs.ErrJobNotFound
					},
				}

				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInvalidState.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], updateJobAction, audit.Attempted, common.Params{jobIDKey: "12345"})
				testapi.VerifyAuditorCalls(calls[1], updateJobAction, audit.Unsuccessful, common.Params{jobIDKey: "12345"})
			})
		})

		Convey("When the job does not exist", func() {
			Convey("Then return status not found (404)", func() {
				reader := strings.NewReader("{\"state\":\"created\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return errs.ErrJobNotFound
					},
				}

				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrJobNotFound.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], updateJobAction, audit.Attempted, common.Params{jobIDKey: "12345"})
				testapi.VerifyAuditorCalls(calls[1], updateJobAction, audit.Unsuccessful, common.Params{jobIDKey: "12345"})
			})
		})

		Convey("When the import api is unable to connect to its datastore", func() {
			Convey("Then return status internal server error (500)", func() {
				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return errs.ErrInternalServer
					},
				}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreInternalError, mockJobService, auditorMock)

				reader := strings.NewReader("{\"state\":\"created\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], updateJobAction, audit.Attempted, common.Params{jobIDKey: "12345"})
				testapi.VerifyAuditorCalls(calls[1], updateJobAction, audit.Unsuccessful, common.Params{jobIDKey: "12345"})
			})
		})

		Convey("When auditing attempted action errors", func() {
			Convey("Then return status internal server error (500)", func() {
				newAuditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						return errors.New("broken")
					},
				}
				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return nil
					},
				}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, newAuditorMock)

				reader := strings.NewReader("{\"state\":\"created\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := newAuditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 1)
				testapi.VerifyAuditorCalls(calls[0], updateJobAction, audit.Attempted, common.Params{jobIDKey: "12345"})

				So(len(mockJobService.UpdateJobCalls()), ShouldEqual, 0)
			})
		})

		Convey(`When the import api is unable to connect to its datastore and
       auditing unsuccessful action errors`, func() {
			Convey("Then return status internal server error (500)", func() {
				newAuditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Unsuccessful {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return errors.New("update job error")
					},
				}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, newAuditorMock)

				reader := strings.NewReader("{\"state\":\"created\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := newAuditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], updateJobAction, audit.Attempted, common.Params{jobIDKey: "12345"})
				testapi.VerifyAuditorCalls(calls[1], updateJobAction, audit.Unsuccessful, common.Params{jobIDKey: "12345"})
			})
		})
	})
}

func TestSuccessfullyUpdateJobState(t *testing.T) {
	t.Parallel()
	Convey("Given a request to update job state", t, func() {
		Convey("When successfully updating job state", func() {
			Convey("Then return status ok (200)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return nil
					},
				}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"state\":\"completed\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusOK)

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], updateJobAction, audit.Attempted, common.Params{jobIDKey: "12345"})
				testapi.VerifyAuditorCalls(calls[1], updateJobAction, audit.Successful, common.Params{jobIDKey: "12345"})
			})
		})

		Convey("When successfully updating job state but auditing action fails", func() {
			Convey("Then return status ok (200)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Successful {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return nil
					},
				}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"state\":\"submitted\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], updateJobAction, audit.Attempted, common.Params{jobIDKey: "12345"})
				testapi.VerifyAuditorCalls(calls[1], updateJobAction, audit.Successful, common.Params{jobIDKey: "12345"})
			})
		})
	})
}