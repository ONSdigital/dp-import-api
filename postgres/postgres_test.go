package postgres

import (
	"database/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/ONSdigital/dp-import-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	createJobSQL           = "INSERT INTO Jobs"
	updateJobStateSQL      = "UPDATE Jobs SET job = jsonb_set"
	createInstanceSQL      = "INSERT INTO Instances"
	findInstanceSQL        = "SELECT instance FROM Instances WHERE"
	addFileToJobSQL        = "UPDATE Jobs SET job = jsonb_set"
	addEventSQL            = "UPDATE Instances SET instance = jsonb_set"
	addDimensionSQL        = "INSERT INTO Dimensions"
	findDimensionsSQL      = "SELECT nodeName, value, nodeId"
	updateDimensionSQL     = "UPDATE Dimensions SET nodeId"
	buildPublishDatasetSQL = "SELECT job->'recipe', job->'s3Files', STRING_AGG"
)

func TestNewPostgresDatastore(t *testing.T) {
	t.Parallel()
	Convey("When creating a postgres datastore no errors are returned", t, func() {
		_, db := NewSQLMockWithSQLStatements()
		_, err := NewDatastore(db)
		So(err, ShouldBeNil)

	})
}

func TestAddJobReturnsJobInstance(t *testing.T) {
	t.Parallel()
	Convey("When creating a new job, a job instance is returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(createJobSQL).WillReturnRows(sqlmock.NewRows([]string{"jobId"}).
			AddRow("123"))
		mock.ExpectQuery(createInstanceSQL).WillReturnRows(sqlmock.NewRows([]string{"InstanceID"}).
			AddRow("321"))
		jobInstance, err := ds.AddJob(&models.NewJob{Recipe: "test", Datasets: []string{"RPI"}})
		So(err, ShouldBeNil)
		So(jobInstance.JobID, ShouldEqual, "123")
		So(jobInstance.InstanceIds, ShouldContain, "321")
	})
}

func TestAddInstanceReturnsInstanceId(t *testing.T) {
	t.Parallel()
	Convey("When creating a new import job, an instanceId is returned", t, func() {
		expectID := "000001"
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare(createInstanceSQL).ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"InstanceID"}).
			AddRow(expectID))
		jobID, err := ds.AddInstance("123", "123")
		So(err, ShouldBeNil)
		So(jobID, ShouldEqual, expectID)
	})
}

func TestGetInstance(t *testing.T) {
	t.Parallel()
	Convey("When an instanceId is provided, the import job state is returned", t, func() {
		jsonContent := "{ \"dataset\":\"123\"  }"
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare(findInstanceSQL).ExpectQuery().
			WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"job"}).
			AddRow(jsonContent))
		state, err := ds.GetInstance("any")
		So(err, ShouldBeNil)
		So(state.Dataset, ShouldEqual, "123")
	})
}

func TestAddEvent(t *testing.T) {
	t.Parallel()
	Convey("When adding an event, no error is returned", t, func() {
		jsonContent := "{ \"dataset\":\"123\"  }"
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare(addEventSQL).ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"job"}).AddRow(jsonContent))
		dataStoreErr := ds.AddEvent("123", &models.Event{Type: "type", Message: "321", Time: "000", MessageOffset: "0001"})
		So(dataStoreErr, ShouldBeNil)
	})
}

func TestGetDimensions(t *testing.T) {
	t.Parallel()
	Convey("When adding a dimension, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare(findInstanceSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"job"}).AddRow("{}"))
		mock.ExpectPrepare(findDimensionsSQL).ExpectQuery().
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"nodeName", "value", "nodeId"}).
				AddRow("node1", "0", "1").AddRow("node2", "2", "2"))
		dimensions, dataStoreErr := ds.GetDimension("123")
		So(dataStoreErr, ShouldBeNil)
		So(len(dimensions), ShouldEqual, 2)
	})
}

func TestAddNodeId(t *testing.T) {
	t.Parallel()
	Convey("When adding a node id, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare(updateDimensionSQL).ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instanceId"}).AddRow("123"))
		dataStoreErr := ds.AddNodeID("123", "node1", &models.Dimension{NodeID: "123"})
		So(dataStoreErr, ShouldBeNil)
	})
}

func TestAddDimension(t *testing.T) {
	t.Parallel()
	Convey("When adding a dimension, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare(findInstanceSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"job"}).AddRow("{}"))
		mock.ExpectPrepare(addDimensionSQL).ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{}))
		dataStoreErr := ds.AddDimension("123", &models.Dimension{NodeName: "name", Value: "123"})
		So(dataStoreErr, ShouldBeNil)
	})
}

func TestUploadFile(t *testing.T) {
	t.Parallel()
	Convey("When adding an uploaded file, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare(addFileToJobSQL).ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instanceId"}).AddRow("123"))
		dataStoreErr := ds.AddUploadedFile("123", &models.UploadedFile{"test1", "s3://aws/bucket/test.xls"})
		So(dataStoreErr, ShouldBeNil)
	})
}

func TestAddUpdateState(t *testing.T) {
	t.Parallel()
	Convey("When updating the job state, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare(updateJobStateSQL).ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instanceId"}).AddRow("123"))
		dataStoreErr := ds.UpdateJobState("123", &models.JobState{State: "Start"})
		So(dataStoreErr, ShouldBeNil)
	})
}

func TestBuildPublishDatasetMessage(t *testing.T) {
	t.Parallel()
	Convey("When building a publish data message, it returns no errors and a struct", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(buildPublishDatasetSQL).WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"recipe", "files", "instanceIds"}).
				AddRow("test", "[{ \"aliasName\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}]", "1,2,3"))
		message, dataStoreError := ds.BuildPublishDatasetMessage("123")
		So(dataStoreError, ShouldBeNil)
		So("test", ShouldEqual, message.Recipe)
		So(1, ShouldEqual, len(message.UploadedFiles))
		So(3, ShouldEqual, len(message.InstanceIds))
	})
}

func NewSQLMockWithSQLStatements() (sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	So(err, ShouldBeNil)
	mock.ExpectBegin()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectPrepare(createJobSQL)
	mock.ExpectPrepare(updateJobStateSQL)
	mock.ExpectPrepare(createInstanceSQL)
	mock.ExpectPrepare(findInstanceSQL)
	mock.ExpectPrepare(addFileToJobSQL)
	mock.ExpectPrepare(addEventSQL)
	mock.ExpectPrepare(addDimensionSQL)
	mock.ExpectPrepare(findDimensionsSQL)
	mock.ExpectPrepare(updateDimensionSQL)
	mock.ExpectPrepare(buildPublishDatasetSQL)
	_, dbError := db.Begin()
	So(dbError, ShouldBeNil)
	return mock, db
}
