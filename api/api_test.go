package api

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/ONSdigital/dp-import-api/mocks"
)

func TestAddJobReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("When a no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader("{ \"NumberOfInstances\": 1, \"recipe\":\"test\"}")
		r, err := http.NewRequest("POST", "http://localhost:21800/job", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{InternalError: true}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})
}

func TestAddJobReturnsBadClientRequest(t *testing.T) {
	t.Parallel()
	Convey("When a empt json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{ }")
		r, err := http.NewRequest("POST", "http://localhost:21800/job", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func TestAddJob(t *testing.T) {
	t.Parallel()
	Convey("When a valid message is sent, a jobInstance model is returned", t, func() {
		reader := strings.NewReader("{ \"NumberOfInstances\": 1, \"recipe\":\"test\"}")
		r, err := http.NewRequest("POST", "http://localhost:21800/job", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Body.String(), ShouldContainSubstring, "\"jobId\":\"34534543543\"")
	})
}

func TestGetInstanceReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When a get request for an instance has an invalid instanceId, return a not found code", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/instance/12345", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestGetImportJobReturnsImportJob(t *testing.T) {
	t.Parallel()
	Convey("When a get request for an import job has a valid instance id, it state is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/instance/12345", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAddS3FileReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When adding a S3 file to an import job with a invalid instance id, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"aliasName\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/job/12345/s3file", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestGetDimensionsReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When a get request for a list of dimensions with an invalid instance id, returns a not found code", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/instance/12345/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestGetDimensionsReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("When a get request for a list of dimensions it returns an Ok code", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/instance/12345/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAddEventReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When adding an event into an instance with an invalid instanceId, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"type\":\"info\",\"message\":\"123 123\",\"time\":\"7789789\",\"messageOffset\":\"321\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/instance/12345/events", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestAddEventReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("When adding an event into an instance with a valid instanceId, it returns an OK code", t, func() {
		reader := strings.NewReader("{ \"type\":\"info\",\"message\":\"123 123\",\"time\":\"7789789\",\"messageOffset\":\"321\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/instance/12345/events", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAddDimensionReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When adding a dimension with an invalid instanceId, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"nodeName\":\"321\",\"value\":\"123 123\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/instance/12345/dimensions", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestAddDimensionReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("When adding a dimension with a valid instanceId, it returns an OK code", t, func() {
		reader := strings.NewReader("{ \"nodeName\":\"321\",\"value\":\"123 123\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/instance/12345/dimensions", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAddNodeIdReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When adding a node id for a dimension with an invalid instanceId, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"nodeId\":\"321\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/instance/12345/dimensions/nodename/nodeId", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestAddNodeIdReturnsOk(t *testing.T) {
	t.Parallel()
	Convey("When adding a node id for a dimension with a valid instanceId, it returns an OK code", t, func() {
		reader := strings.NewReader("{ \"nodeId\":\"321\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/instance/12345/dimensions/nodename/nodeId", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestUpdateJobState(t *testing.T) {
	t.Parallel()
	Convey("When updating a jobs state, it returns an OK code", t, func() {
		reader := strings.NewReader("{ \"state\":\"start\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/job/12345/state", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{}, make(chan []byte, 1))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestUpdateJobStateReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When updating a jobs state with an invalid jobId, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"state\":\"start\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/job/12345/state", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(&mocks.DataStore{NotFound: true}, make(chan []byte))
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestUpdateJobStateToSubmitted(t *testing.T) {
	t.Parallel()
	Convey("When updating a jobs state with an invalid jobId, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"state\":\"submitted\"}")
		r, err := http.NewRequest("PUT", "http://localhost:21800/job/12345/state", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		producer := make(chan []byte, 1)
		api := CreateImportAPI(&mocks.DataStore{}, producer)
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}
