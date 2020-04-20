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

func TestFailureToGetJobs(t *testing.T) {
	t.Parallel()

	attemptedAuditParams := common.Params{"caller_identity": "someone@ons.gov.uk"}
	p := common.Params{}

	Convey("Given a request to get a list of jobs", t, func() {
		Convey("When no auth token is provided", func() {
			Convey("Then return status unauthorised (401)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{}
				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock, nil)

				r, err := testapi.CreateRequestWithOutAuth("GET", "http://localhost:21800/jobs", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnauthorised.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], getJobsAction, audit.Attempted, p)
				testapi.VerifyAuditorCalls(calls[1], getJobsAction, audit.Unsuccessful, p)
			})
		})

		Convey("When there is no available datastore", func() {
			Convey("Then return status internal error (500)", func() {
				mockJobService := &testapi.JobServiceMock{}
				auditorMock := testapi.NewAuditorMock()

				api := routes(mux.NewRouter(), &testapi.DstoreInternalError, mockJobService, auditorMock, nil)

				w := httptest.NewRecorder()
				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
				So(err, ShouldBeNil)

				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], getJobsAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], getJobsAction, audit.Unsuccessful, p)
			})
		})

		Convey("When request is valid but auditing action attempted fails", func() {
			Convey("Then return status internal server error (500)", func() {
				mockJobService := &testapi.JobServiceMock{}
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Attempted {
							return errors.New("BOOM")
						}
						return nil
					},
				}
				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock, nil)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()

				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 1)

				testapi.VerifyAuditorCalls(calls[0], getJobsAction, audit.Attempted, attemptedAuditParams)
			})
		})

		Convey(`When the import api is unable to connect to its datastore and the
      auditing of the unsuccessful action errors`, func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if action == getJobsAction && result == audit.Unsuccessful {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := routes(mux.NewRouter(), &testapi.DstoreInternalError, mockJobService, auditorMock, nil)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], getJobsAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], getJobsAction, audit.Unsuccessful, p)
			})
		})

		Convey(`When retrieval of resources from datastore is successful but
        auditing action successful fails`, func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Successful {
							return errors.New("BOOM")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock, nil)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], getJobsAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], getJobsAction, audit.Successful, p)
			})
		})
	})
}

func TestGetJobs(t *testing.T) {
	t.Parallel()

	attemptedAuditParams := common.Params{"caller_identity": "someone@ons.gov.uk"}

	Convey("Given a request to get a list of jobs", t, func() {
		Convey("When retrieval of resources from datastore is successful", func() {
			Convey("Then return status ok (200)", func() {
				mockJobService := &testapi.JobServiceMock{}
				auditorMock := testapi.NewAuditorMock()

				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock, nil)

				w := httptest.NewRecorder()
				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
				So(err, ShouldBeNil)

				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				p := common.Params{}
				testapi.VerifyAuditorCalls(calls[0], getJobsAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], getJobsAction, audit.Successful, p)
			})
		})
	})

	Convey(`Given a request to get a list of jobs with a filter of state is
	  'completed'`, t, func() {
		Convey("When retrieval of resources from datastore is successful", func() {
			Convey("Then return status ok (200)", func() {
				mockJobService := &testapi.JobServiceMock{}
				auditorMock := testapi.NewAuditorMock()

				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock, nil)

				w := httptest.NewRecorder()
				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs?state=completed", nil)
				So(err, ShouldBeNil)

				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				p := common.Params{"filterQuery": "completed"}
				testapi.VerifyAuditorCalls(calls[0], getJobsAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], getJobsAction, audit.Successful, p)
			})
		})
	})
}
