package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-import-api/api/testapi"
	"github.com/ONSdigital/dp-import-api/models"
	testmongo "github.com/ONSdigital/dp-import-api/mongo/testmongo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIncreaseProcessedInstanceHandler(t *testing.T) {

	t.Parallel()

	Convey("Given a request to increase the processed count for an instance", t, func() {
		w := httptest.NewRecorder()
		r, err := testapi.CreateRequestWithAuth(http.MethodPut, "http://localhost:21800/jobs/34534543543/processed/54321", nil)
		So(err, ShouldBeNil)

		Convey("When the update is successful", func() {
			ds := &testmongo.DataStorer{}
			api := SetupAPIWith(ds, nil)
			api.router.ServeHTTP(w, r)

			Convey("Then the returned status code 200 OK, with the expected body", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				validateBody(w.Body, []models.ProcessedInstances{
					{
						ID:             "54321",
						RequiredCount:  5,
						ProcessedCount: 1,
					},
				})
			})

			Convey("Then the datastore has been locked, but is no longer locked", func() {
				So(ds.HasBeenLocked, ShouldBeTrue)
				So(ds.IsLocked, ShouldBeFalse)
			})
		})

		Convey("When the datastore returns an InternalError", func() {
			api := SetupAPIWith(&testapi.DstoreInternalError, nil)
			api.router.ServeHTTP(w, r)

			Convey("Then the returned status code 500 Internal Server Error, with the expected body", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldEqual, "internal error\n")
			})
		})

		Convey("When the instance does not exist for the import job", func() {
			r.URL.Path = "/jobs/34534543543/processed/inexistent"
			api := SetupAPIWith(nil, nil)
			api.router.ServeHTTP(w, r)

			Convey("Then the returned status code 400 Bad request, with the expected body", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldEqual, "the instance id was not found in the provided job\n")
			})
		})
	})
}

func validateBody(b *bytes.Buffer, expected []models.ProcessedInstances) {
	actual := []models.ProcessedInstances{}
	err := json.Unmarshal(b.Bytes(), &actual)
	So(err, ShouldBeNil)
	So(actual, ShouldResemble, expected)
}
