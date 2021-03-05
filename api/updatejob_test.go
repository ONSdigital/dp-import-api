package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-import-api/api/testapi"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFailureToUpdateJobState(t *testing.T) {
	t.Parallel()

	Convey("Given a request to update job state", t, func() {
		Convey("When request has no auth header", func() {
			Convey("Then return status unauthorised (401)", func() {
				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return nil
					},
				}
				api := SetupAPIWith(nil, mockJobService)

				reader := strings.NewReader("{\"state\":\"created\"}")
				r, err := testapi.CreateRequestWithOutAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnauthorised.Error())

				Convey("Then the request body has been drained", func() {
					bytesRead, err := r.Body.Read(make([]byte, 1))
					So(bytesRead, ShouldEqual, 0)
					So(err, ShouldEqual, io.EOF)
				})
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

				api := SetupAPIWith(&testapi.DstoreNotFound, mockJobService)
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrFailedToParseJSONBody.Error())

				Convey("Then the request body has been drained", func() {
					bytesRead, err := r.Body.Read(make([]byte, 1))
					So(bytesRead, ShouldEqual, 0)
					So(err, ShouldEqual, io.EOF)
				})
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

				api := SetupAPIWith(&testapi.DstoreNotFound, mockJobService)
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInvalidState.Error())

				Convey("Then the request body has been drained", func() {
					bytesRead, err := r.Body.Read(make([]byte, 1))
					So(bytesRead, ShouldEqual, 0)
					So(err, ShouldEqual, io.EOF)
				})
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

				api := SetupAPIWith(&testapi.DstoreNotFound, mockJobService)
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrJobNotFound.Error())

				Convey("Then the request body has been drained", func() {
					bytesRead, err := r.Body.Read(make([]byte, 1))
					So(bytesRead, ShouldEqual, 0)
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey("When the import api is unable to connect to its datastore", func() {
			Convey("Then return status internal server error (500)", func() {
				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return errs.ErrInternalServer
					},
				}
				api := SetupAPIWith(&testapi.DstoreInternalError, mockJobService)

				reader := strings.NewReader("{\"state\":\"created\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				Convey("Then the request body has been drained", func() {
					bytesRead, err := r.Body.Read(make([]byte, 1))
					So(bytesRead, ShouldEqual, 0)
					So(err, ShouldEqual, io.EOF)
				})
			})
		})
	})
}

func TestSuccessfullyUpdateJobState(t *testing.T) {
	t.Parallel()

	Convey("Given a request to update job state", t, func() {
		Convey("When successfully updating job state", func() {
			Convey("Then return status ok (200)", func() {
				mockJobService := &testapi.JobServiceMock{
					UpdateJobFunc: func(ctx context.Context, jobID string, job *models.Job) error {
						return nil
					},
				}
				api := SetupAPIWith(nil, mockJobService)

				reader := strings.NewReader("{ \"state\":\"completed\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusOK)

				Convey("Then the request body has been drained", func() {
					bytesRead, err := r.Body.Read(make([]byte, 1))
					So(bytesRead, ShouldEqual, 0)
					So(err, ShouldEqual, io.EOF)
				})
			})
		})
	})
}
