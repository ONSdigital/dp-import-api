package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-import-api/api/testapi"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFailureToGetJobs(t *testing.T) {
	t.Parallel()

	Convey("Given a request to get a list of jobs", t, func() {
		Convey("When no auth token is provided", func() {
			Convey("Then return status unauthorised (401)", func() {
				api := SetupAPIWith(nil, nil)

				r, err := testapi.CreateRequestWithOutAuth("GET", "http://localhost:21800/jobs", nil)
				So(err, ShouldBeNil)

				w := httptest.NewRecorder()
				api.router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnauthorised.Error())
			})
		})

		Convey("When there is no available datastore", func() {
			Convey("Then return status internal error (500)", func() {

				api := SetupAPIWith(&testapi.DstoreInternalError, nil)

				w := httptest.NewRecorder()
				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
				So(err, ShouldBeNil)

				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
			})
		})
	})
}

func TestGetJobs(t *testing.T) {
	t.Parallel()

	Convey("Given a request to get a list of jobs", t, func() {
		Convey("When retrieval of resources from datastore is successful", func() {
			Convey("Then return status ok (200)", func() {
				api := SetupAPIWith(nil, nil)

				w := httptest.NewRecorder()
				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
				So(err, ShouldBeNil)

				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})

	Convey(`Given a request to get a list of jobs with a filter of state is
	  'completed'`, t, func() {
		Convey("When retrieval of resources from datastore is successful", func() {
			Convey("Then return status ok (200)", func() {
				api := SetupAPIWith(nil, nil)
				w := httptest.NewRecorder()
				r, err := testapi.CreateRequestWithAuth("GET", "http://localhost:21800/jobs?state=completed", nil)
				So(err, ShouldBeNil)

				api.router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
