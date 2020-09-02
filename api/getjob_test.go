package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-import-api/api/testapi"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFailureToGetJob(t *testing.T) {
	t.Parallel()

	Convey("Given a request to get a job", t, func() {
		Convey("When no auth token is provided", func() {
			Convey("Then return status unauthorised (401)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, nil)

				r, err := testapi.CreateRequestWithOutAuth("GET", "http://localhost:21800/jobs/123", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnauthorised.Error())
			})
		})

		Convey("When the request contains an invalid jobID", func() {
			Convey("Then return status not found (404)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := routes(mux.NewRouter(), &testapi.DstoreNotFound, mockJobService, nil)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrJobNotFound.Error())
			})
		})
	})
}

func TestSuccessfullyGetJob(t *testing.T) {
	t.Parallel()

	Convey("Given a request to get a job", t, func() {
		Convey("When retrieval of job from datastore is successful", func() {
			Convey("Then return status ok (200)", func() {
				mockJobService := &testapi.JobServiceMock{}
				api := routes(mux.NewRouter(), &testapi.Dstore, mockJobService, nil)

				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
