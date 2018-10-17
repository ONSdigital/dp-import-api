package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-import-api/api/testapi"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFailureToGetJob(t *testing.T) {
	t.Parallel()

	attemptedAuditParams := common.Params{"caller_identity": "someone@ons.gov.uk", "job_id": "123"}

	Convey("Given a request to get a job", t, func() {
		Convey("When no auth token is provided", func() {
			Convey("Then return status unauthorised (401)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				r, err := testapi.CreateRequestWithOutAuth("GET", "http://localhost:21800/jobs/123", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnauthorised.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], getJobAction, audit.Attempted, common.Params{jobIDKey: "123"})
				testapi.VerifyAuditorCalls(calls[1], getJobAction, audit.Unsuccessful, common.Params{jobIDKey: "123"})
			})
		})

		Convey("When the request contains an invalid jobID", func() {
			Convey("Then return status not found (404)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrJobNotFound.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], getJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], getJobAction, audit.Unsuccessful, common.Params{jobIDKey: "123"})
			})
		})

		Convey("When auditing attempted action errors", func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if action == getJobAction && result == audit.Attempted {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 1)
				testapi.VerifyAuditorCalls(calls[0], getJobAction, audit.Attempted, attemptedAuditParams)
			})
		})

		Convey(`When the import api is unable to connect to its datastore and the
      auditing of the unsuccessful action errors`, func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if action == getJobAction && result == audit.Unsuccessful {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreInternalError, mockJobService, auditorMock)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], getJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], getJobAction, audit.Unsuccessful, common.Params{jobIDKey: "123"})
			})
		})

		Convey(`When retrieval of job from datastore is successful but auditing
      action successful errors`, func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if action == getJobAction && result == audit.Successful {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], getJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], getJobAction, audit.Successful, common.Params{jobIDKey: "123"})
			})
		})
	})
}

func TestSuccessfullyGetJob(t *testing.T) {
	t.Parallel()

	attemptedAuditParams := common.Params{"caller_identity": "someone@ons.gov.uk", "job_id": "123"}

	Convey("Given a request to get a job", t, func() {
		Convey("When retrieval of job from datastore is successful", func() {
			Convey("Then return status ok (200)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusOK)

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], getJobAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], getJobAction, audit.Successful, common.Params{jobIDKey: "123"})
			})
		})
	})
}
