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
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFailureToAddFile(t *testing.T) {
	t.Parallel()

	Convey("Given a request to add a s3 file", t, func() {
		Convey("When no auth token is provided", func() {
			Convey("Then return status unauthorised (401)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := Setup(mux.NewRouter(), &testapi.Dstore, mockJobService)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithOutAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnauthorised.Error())
			})
		})

		Convey("When the request body is invalid", func() {
			Convey("Then return status bad request (400)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := Setup(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService)

				reader := strings.NewReader("{")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrFailedToParseJSONBody.Error())
			})
		})

		Convey("When the request body is missing a mandatory field", func() {
			Convey("Then return status bad request (400)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := Setup(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService)

				reader := strings.NewReader("{ \"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInvalidUploadedFileObject.Error())
			})
		})

		Convey("When the import api is unable to connect to its datastore", func() {
			Convey("Then return status internal server error (500)", func() {
				mockJobService := &testapi.JobServiceMock{
					CreateJobFunc: func(ctx context.Context, newJob *models.Job) (*models.Job, error) {
						return nil, errs.ErrInternalServer
					},
				}
				api := Setup(mux.NewRouter(), &testapi.DstoreInternalError, mockJobService)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
			})
		})

		Convey("When the request contains an invalid 'instance_id'", func() {
			Convey("Then return status not found (404)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := Setup(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrJobNotFound.Error())
			})
		})
	})
}

func TestSuccessfullyAddFile(t *testing.T) {
	t.Parallel()

	Convey("Given a request to add a s3 file", t, func() {
		Convey("When request is valid", func() {
			Convey("Then a retuen status ok (200)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := Setup(mux.NewRouter(), &testapi.Dstore, mockJobService)

				reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
				r, err := testapi.CreateRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
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
