package postgres

import (
	"database/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/ONSdigital/dp-import-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	createJobSQL           = "INSERT INTO Jobs"
	getJobSQL              = "SELECT instanceId, job FROM Jobs "
	getJobsSQL             = "SELECT Jobs.jobId, instanceId"
	updateJobStateSQL      = "UPDATE Jobs set job = job"
	addFileToJobSQL        = "UPDATE Jobs SET job = jsonb_set"
	createInstanceSQL      = "INSERT INTO Instances"
	findInstanceSQL        = "SELECT instance, jobId FROM Instances WHERE"
	updateInstanceSQL      = "UPDATE Instances set instance = instance"
	addEventSQL            = "UPDATE Instances SET instance = jsonb_set"
	addDimensionSQL        = "INSERT INTO Dimensions"
	findDimensionsSQL      = "SELECT dimensionName, value, nodeId"
	getDimensionValuesSQL = "SELECT dimensions.value FROM dimensions"
	updateDimensionSQL     = "UPDATE Dimensions SET nodeId"
	buildPublishDatasetSQL = "SELECT job->>'recipe', job->'files', STRING_AGG"
)

// go-sqlmock libray does not support all transations methods (eg tx.Stmt(*).Query(...)). So
// AddJob and AddInstance functions do not have tests.

func TestNewPostgresDatastore(t *testing.T) {
	t.Parallel()
	Convey("When creating a postgres datastore no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		_, err := NewDatastore(db)
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetInstance(t *testing.T) {
	t.Parallel()
	Convey("When an instanceId is provided, the importqueue job state is returned", t, func() {
		jsonContent := "{ \"state\":\"Created\"}"
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(findInstanceSQL).WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"jobId","instance"}).
			AddRow( jsonContent, "1"))
		state, err := ds.GetInstance("http://localhost:80", "any")
		So(err, ShouldBeNil)
		So(state.State, ShouldEqual, "Created")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetJobs(t *testing.T) {
	t.Parallel()
	Convey("When get jobs is called, a list of jobs are returned", t, func() {
		jsonContent := "{ \"state\":\"Created\"}"
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(getJobsSQL).
			WillReturnRows(sqlmock.NewRows([]string{"jobid", "instanceid","json"}).
			AddRow(1, 1, jsonContent))
		jobs, err := ds.GetJobs("localhost", []string{})
		So(err, ShouldBeNil)
		So(jobs[0].State, ShouldEqual, "Created")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetJob(t *testing.T) {
	t.Parallel()
	Convey("When get jobs is called, a list of jobs are returned", t, func() {
		jsonContent := "{ \"state\":\"Created\"}"
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(getJobSQL).
			WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"instanceid","json"}).
			AddRow( 1, jsonContent))
		state, err := ds.GetJob("localhost", "123")
		So(err, ShouldBeNil)
		So(state.State, ShouldEqual, "Created")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}


func TestAddEvent(t *testing.T) {
	t.Parallel()
	Convey("When adding an event, no error is returned", t, func() {
		jsonContent := "{ \"dataset\":\"123\"  }"
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(addEventSQL).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"job"}).AddRow(jsonContent))
		dataStoreErr := ds.AddEvent("123", &models.Event{Type: "type", Message: "321", Time: "000", MessageOffset: "0001"})
		So(dataStoreErr, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetDimensions(t *testing.T) {
	t.Parallel()
	Convey("When adding a dimension, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(findInstanceSQL).WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instance", "jobId"}).AddRow("{}", "1"))
		mock.ExpectQuery(findDimensionsSQL).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"nodeName", "value", "nodeId"}).
				AddRow("node1", "0", "1").AddRow("node2", "2", "2"))
		dimensions, dataStoreErr := ds.GetDimensions("123")
		So(dataStoreErr, ShouldBeNil)
		So(len(dimensions), ShouldEqual, 2)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestAddNodeId(t *testing.T) {
	t.Parallel()
	Convey("When adding a node id, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(updateDimensionSQL).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instanceId"}).AddRow("123"))
		dataStoreErr := ds.AddNodeID("123", "node1", &models.Dimension{NodeID: "123"})
		So(dataStoreErr, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestAddDimension(t *testing.T) {
	t.Parallel()
	Convey("When adding a dimension, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(findInstanceSQL).WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instance", "jobId"}).AddRow("{}", "1"))
		mock.ExpectQuery(addDimensionSQL).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{}))
		dataStoreErr := ds.AddDimension("123", &models.Dimension{Name: "name", Value: "123"})
		So(dataStoreErr, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetDimensionValues(t *testing.T) {
	t.Parallel()
	Convey("When getting a list of dimension values, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(getDimensionValuesSQL).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"values"}).AddRow("35").AddRow("45"))
		dimension, err := ds.GetDimensionValues("123","age")
		So(err, ShouldBeNil)
		So(dimension.Values, ShouldContain, "35")
		So(dimension.Values, ShouldContain, "45")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestUploadFile(t *testing.T) {
	t.Parallel()
	Convey("When adding an uploaded file, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(addFileToJobSQL).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instanceId"}).AddRow("123"))
		dataStoreErr := ds.AddUploadedFile("123", &models.UploadedFile{"test1", "s3://aws/bucket/test.xls"})
		So(dataStoreErr, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestUpdateJobState(t *testing.T) {
	t.Parallel()
	Convey("When updating the job state, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(updateJobStateSQL).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instanceId"}).AddRow("123"))
		dataStoreErr := ds.UpdateJobState("123", &models.Job{State: "Start"})
		So(dataStoreErr, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestUpdateInstanceState(t *testing.T) {
	t.Parallel()
	Convey("When updating the instance state, no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(updateInstanceSQL).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instanceId"}).AddRow("123"))
		dataStoreErr := ds.UpdateInstance("123", &models.Instance{NumberOfObservations: 5})
		So(dataStoreErr, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
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
		message, dataStoreError := ds.BuildImportDataMessage("123")
		So(dataStoreError, ShouldBeNil)
		So("test", ShouldEqual, message.Recipe)
		So(1, ShouldEqual, len(message.UploadedFiles))
		So(3, ShouldEqual, len(message.InstanceIDs))
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func NewSQLMockWithSQLStatements() (sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	So(err, ShouldBeNil)
	mock.ExpectBegin()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectPrepare(createJobSQL)
	mock.ExpectPrepare(getJobSQL)
	mock.ExpectPrepare(getJobsSQL)
	mock.ExpectPrepare(updateJobStateSQL)
	mock.ExpectPrepare(addFileToJobSQL)
	mock.ExpectPrepare(createInstanceSQL)
	mock.ExpectPrepare(updateInstanceSQL)
	mock.ExpectPrepare(findInstanceSQL)
	mock.ExpectPrepare(addEventSQL)
	mock.ExpectPrepare(addDimensionSQL)
	mock.ExpectPrepare(findDimensionsSQL)
	mock.ExpectPrepare(getDimensionValuesSQL)
	mock.ExpectPrepare(updateDimensionSQL)
	mock.ExpectPrepare(buildPublishDatasetSQL)
	_, dbError := db.Begin()
	So(dbError, ShouldBeNil)
	return mock, db
}