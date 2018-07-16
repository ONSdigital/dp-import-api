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

func TestFailureToAddFile(t *testing.T) {

	t.Parallel()
	attemptedAuditParams := common.Params{"caller_identity": "someone@ons.gov.uk", "job_id": "12345"}

	Convey("Given a request to add a s3 file", t, func() {
		Convey("When no auth token is provided", func() {
			Convey("Then return status unauthorised (401)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithOutAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnauthorised.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, common.Params{"job_id": "12345"})
				testapi.VerifyAuditorCalls(calls[1], uploadFileAction, audit.Unsuccessful, common.Params{"job_id": "12345"})
			})
		})

		Convey("When the request body is invalid", func() {
			Convey("Then return status bad request (400)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)

				reader := strings.NewReader("{")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrFailedToParseJSONBody.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], uploadFileAction, audit.Unsuccessful, common.Params{"job_id": "12345"})
			})
		})

		Convey("When the request body is missing a mandatory field", func() {
			Convey("Then return status bad request (400)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInvalidUploadedFileObject.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], uploadFileAction, audit.Unsuccessful, common.Params{"job_id": "12345"})
			})
		})

		params := common.Params{
			jobIDKey:    "12345",
			"fileAlias": "n1",
			"fileURL":   "https://aws.s3/ons/myfile.exel",
		}

		Convey("When the import api is unable to connect to its datastore", func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, newJob *models.Job) (*models.Job, error) {
						return nil, errs.ErrInternalServer
					},
				}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreInternalError, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], uploadFileAction, audit.Unsuccessful, params)
			})
		})

		Convey("When the request contains an invalid 'instance_id'", func() {
			Convey("Then return status not found (404)", func() {
				auditorMock := testapi.NewAuditorMock()
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrJobNotFound.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)

				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], uploadFileAction, audit.Unsuccessful, params)
			})
		})

		Convey("When auditing attempted action errors", func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Attempted {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 1)

				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, attemptedAuditParams)
			})
		})

		Convey(`When the request body is missing a mandatory field and the auditing
      of the unsuccessful action errors`, func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Unsuccessful {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], uploadFileAction, audit.Unsuccessful, common.Params{"job_id": "12345"})
			})
		})

		Convey(`When the import api is unable to connect to its datastore and the
      auditing of the unsuccessful action errors`, func() {
			Convey("Then return status internal server error (500)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Unsuccessful {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], uploadFileAction, audit.Unsuccessful, params)
			})
		})
	})
}

func TestSuccessfullyAddFile(t *testing.T) {
	t.Parallel()

	attemptedAuditParams := common.Params{"caller_identity": "someone@ons.gov.uk", "job_id": "12345"}

	params := common.Params{
		jobIDKey:    "12345",
		"fileAlias": "n1",
		"fileURL":   "https://aws.s3/ons/myfile.exel",
	}

	Convey("Given a request to add a s3 file", t, func() {
		Convey("When request is valid", func() {
			Convey("Then a retuen status ok (200)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusOK)

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], uploadFileAction, audit.Successful, params)

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey("When upload is successful but auditing action successful errors", func() {
			Convey("Then return status ok (200)", func() {
				auditorMock := &audit.AuditorServiceMock{
					RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
						if result == audit.Successful {
							return errors.New("audit error")
						}
						return nil
					},
				}
				mockJobService := &testapi.JobServiceMock{}
				api := CreateImportAPI(mux.NewRouter(), &testapi.Dstore, mockJobService, auditorMock)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusOK)

				calls := auditorMock.RecordCalls()
				So(len(calls), ShouldEqual, 2)
				testapi.VerifyAuditorCalls(calls[0], uploadFileAction, audit.Attempted, attemptedAuditParams)
				testapi.VerifyAuditorCalls(calls[1], uploadFileAction, audit.Successful, params)

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})
	})
}
