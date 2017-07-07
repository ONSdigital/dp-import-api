package api

import (
	"github.com/ONSdigital/dp-import-api/mocks/datastore"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddImportJobReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("When a no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader("{ \"datasets\": [\"test123\"], \"recipe\":\"test\"}")
		r, err := http.NewRequest("POST", "http://localhost:21800/job", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{InternalError: true})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})
}

func TestAddImportJobReturnsBadClientRequest(t *testing.T) {
	t.Parallel()
	Convey("When a empt json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{ }")
		r, err := http.NewRequest("POST", "http://localhost:21800/job", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func TestAddImportJobReturnsJobInstance(t *testing.T) {
	t.Parallel()
	Convey("When a valid import job message is sent, an instanceId is returned", t, func() {
		reader := strings.NewReader("{ \"datasets\": [\"test123\"], \"recipe\":\"test\"}")
		r, err := http.NewRequest("POST", "http://localhost:21800/job", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Body.String(), ShouldContainSubstring, "\"jobId\":\"34534543543\"")
	})
}

func TestGetImportJobReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When a get request for an import is job has a invalid instance id, not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/import/12345", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestGetImportJobReturnsImportJob(t *testing.T) {
	t.Parallel()
	Convey("When a get request for an import job has a valid instance id, it state is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/import/12345", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAddS3FileReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When adding a S3 file to an import job with a invalid instance id, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"aliasName\":\"n1\",\"s3Url\":\"https://aws.s3/ons/myfile.exel\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/import/12345/s3file", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestGetDimensionsReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When a get request for a list of dimensions within a import a invalid instance id, returns not found code", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/import/12345/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestGetDimensionsReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("When a get request for a list of dimensions it returns a Ok code", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/import/12345/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAddEventReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When adding an event into a import job with an invalid instanceId, it returns not found code", t, func() {
		reader := strings.NewReader("{ \"type\":\"info\",\"message\":\"123 123\",\"time\":\"7789789\",\"messageOffset\":\"321\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/import/12345/events", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestAddEventReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("When adding an event into a import job with an invalid instanceId, it returns OK code", t, func() {
		reader := strings.NewReader("{ \"type\":\"info\",\"message\":\"123 123\",\"time\":\"7789789\",\"messageOffset\":\"321\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/import/12345/events", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAddDimensionReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When adding a dimension into a import job with an invalid instanceId, it returns not found code", t, func() {
		reader := strings.NewReader("{ \"nodeName\":\"321\",\"value\":\"123 123\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/import/12345/dimensions", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestAddDimensionReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("When adding a dimension into a import job with an invalid instanceId, it returns OK code", t, func() {
		reader := strings.NewReader("{ \"nodeName\":\"321\",\"value\":\"123 123\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/import/12345/dimensions", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAddNodeIdReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When adding a node id into a import job with an invalid instanceId, it returns not found code", t, func() {
		reader := strings.NewReader("{ \"nodeId\":\"321\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/import/12345/dimensions/nodename/nodeId", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestAddNodeIdReturnsNotOk(t *testing.T) {
	t.Parallel()
	Convey("When adding an node id into a import job with an invalid instanceId, it returns OK code", t, func() {
		reader := strings.NewReader("{ \"nodeId\":\"321\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/import/12345/dimensions/nodename/nodeId", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{})
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}
