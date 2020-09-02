package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-import-api/api/testapi"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var dummyJob = &models.Job{ID: "34534543543"}

func TestFailureToAddJob(t *testing.T) {
	t.Parallel()

	Convey("Given a request to add a job", t, func() {
		Convey("When no auth token is provided", func() {
			Convey("Then return status unauthorised (401)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, nil)

				reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
				r, err := testapi.CreateRequestWithOutAuth("POST", "http://localhost:21800/jobs", reader)
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

		Convey("When the request body is invalid", func() {
			Convey("Then return status bad request (400)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := routes(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, nil)

				reader := strings.NewReader("{")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrFailedToParseJSONBody.Error())

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
				api := routes(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, nil)

				reader := strings.NewReader("{ \"number_of_instances\": 1}")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInvalidJob.Error())

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
				api := routes(mux.NewRouter(), &testapi.DstoreInternalError, mockJobService, nil)

				reader := strings.NewReader(`{"recipe":"test"}`)
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey(`When creating the new job a duplication key error occurs but the
			auditing of unsuccessful action fails`, func() {

			Convey("Then return status internal server error (500)", func() {

				reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
						return nil, errors.New("duplication key error")
					},
				}

				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, nil)
				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

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

	Convey("Given a valid request to add a job", t, func() {
		Convey("When successfully created in datastore", func() {
			Convey("Then return status created (201)", func() {
				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
						return dummyJob, nil
					},
				}
				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, nil)

				reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusCreated)
				So(w.Body.String(), ShouldContainSubstring, "\"id\":\"34534543543\"")

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})

		Convey("When auditing request but the successful action fails", func() {
			Convey("Then continue to return status created (201)", func() {
				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, job *models.Job) (*models.Job, error) {
						return dummyJob, nil
					},
				}
				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, nil)

				reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
				r, err := testapi.CreateRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusCreated)
				So(w.Body.String(), ShouldContainSubstring, "\"id\":\"34534543543\"")

				Convey("Then the request body has been drained", func() {
					_, err = r.Body.Read(make([]byte, 1))
					So(err, ShouldEqual, io.EOF)
				})
			})
		})
	})
}
