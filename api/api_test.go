package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-import-api/api/testapi"
	mockdatasetclient "github.com/ONSdigital/dp-import-api/dataset/testapi"
	mockdatastore "github.com/ONSdigital/dp-import-api/mongo/testmongo"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	host                = "http://localhost:21800"
	dsetAPI             = mockdatasetclient.CreateDatasetAPI()
	dstore              = mockdatastore.DataStorer{}
	dstoreNotFound      = mockdatastore.DataStorer{NotFound: true}
	dstoreInternalError = mockdatastore.DataStorer{InternalError: true}
	jimporter           = mockapi.JobImporter{}
)

const secretKey = "123"

func TestAddJobReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("When no datastore is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"recipe":"test"}`)
		r, err := createRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstoreInternalError, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})
}

func TestGetJobsReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("When a get jobs request has no available datastore, an internal error is returned", t, func() {
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstoreInternalError, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})
}

func TestGetJobs(t *testing.T) {
	t.Parallel()
	Convey("When a get jobs request has a datastore, an ok status is returned ", t, func() {
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstore, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestGetJobReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When a get job request has a invalid jobID, a not found status is returned", t, func() {
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs/000000", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstoreNotFound, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestGetJob(t *testing.T) {
	t.Parallel()
	Convey("When a no data store is available, an internal error is returned", t, func() {
		r, err := createRequestWithAuth("GET", "http://localhost:21800/jobs/123", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstore, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAddJobReturnsBadClientRequest(t *testing.T) {
	t.Parallel()
	Convey("When a empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{ }")
		r, err := createRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstoreNotFound, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func TestAddJob(t *testing.T) {
	t.Parallel()
	Convey("When a valid message is sent, a jobInstance model is returned", t, func() {
		reader := strings.NewReader("{ \"number_of_instances\": 1, \"recipe\":\"test\"}")
		r, err := createRequestWithAuth("POST", "http://localhost:21800/jobs", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstore, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(w.Body.String(), ShouldContainSubstring, "\"job_id\":\"34534543543\"")
	})
}

func TestAddS3FileReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When adding a S3 file to an importqueue job with a invalid instance id, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345/files", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstoreNotFound, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestUpdateJobState(t *testing.T) {
	t.Parallel()
	Convey("When updating a jobs state, it returns an OK code", t, func() {
		reader := strings.NewReader("{ \"state\":\"start\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstore, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestUpdateJobStateReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("When updating a jobs state with an invalid jobId, it returns a not found code", t, func() {
		reader := strings.NewReader("{ \"state\":\"start\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstoreNotFound, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestUpdateJobStateToSubmitted(t *testing.T) {
	t.Parallel()
	Convey("When a job state is updated to submitted, a message is sent into the job queue", t, func() {
		reader := strings.NewReader("{ \"state\":\"submitted\"}")
		r, err := createRequestWithAuth("PUT", "http://localhost:21800/jobs/12345", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateImportAPI(host, mux.NewRouter(), &dstore, &jimporter, secretKey, dsetAPI)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func createRequestWithAuth(method, URL string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, URL, body)
	r.Header.Set("internal-token", secretKey)
	return r, err
}
