package datastore

import (
	"database/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/ONSdigital/dp-import-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewPostgresDatastore(t *testing.T) {
	Convey("When a import message has no body, an error is returned", t, func() {
		_, db := NewSQLMockWithSQLStatements()
		defer db.Close()
		_, err := NewPostgresDatastore(db)
		So(err, ShouldBeNil)

	})
}

func TestAddJobReturnsJobInstance(t *testing.T) {
	Convey("When creating a new job, a job instance is returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		db.Begin()
		ds, err := NewPostgresDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery("INSERT INTO Job").WillReturnRows(sqlmock.NewRows([]string{"jobId"}).
			AddRow("123"))
		mock.ExpectQuery("INSERT INTO Instances").WillReturnRows(sqlmock.NewRows([]string{"InstanceId"}).
			AddRow("321"))
		jobInstance, err := ds.AddJob(&models.ImportJob{Recipe: "test", Datasets: []string{"RPI"}})
		So(err, ShouldBeNil)
		So(jobInstance.JobId, ShouldEqual, "123")
		So(jobInstance.InstanceIds, ShouldContain, "321")
	})
}

func TestAddInstanceReturnsInstanceId(t *testing.T) {
	Convey("When creating a new import job, an instanceId is returned", t, func() {
		expectId := "000001"
		mock, db := NewSQLMockWithSQLStatements()
		db.Begin()
		ds, err := NewPostgresDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare("INSERT INTO Instance").ExpectQuery().
			WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"InstanceId"}).
			AddRow(expectId))
		jobId, err := ds.AddInstance("123")
		So(err, ShouldBeNil)
		So(jobId, ShouldEqual, expectId)
	})
}

func TestGetInstance(t *testing.T) {
	Convey("When an instanceId is provided, the import job state is returned", t, func() {
		jsonContent := "{ \"dataset\":\"123\"  }"
		mock, db := NewSQLMockWithSQLStatements()
		db.Begin()
		ds, err := NewPostgresDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare("SELECT instance FROM Instances WHERE").ExpectQuery().
			WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"job"}).
			AddRow(jsonContent))
		state, err := ds.GetInstance("any")
		So(err, ShouldBeNil)
		So(state.Dataset, ShouldEqual, "123")
	})
}

func TestDataStoreAddEvent(t *testing.T) {
	Convey("When adding an event, no error is returned", t, func() {
		jsonContent := "{ \"dataset\":\"123\"  }"
		mock, db := NewSQLMockWithSQLStatements()
		db.Begin()
		ds, err := NewPostgresDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare("UPDATE Instances SET instance = jsonb_set").ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"job"}).AddRow(jsonContent))
		dataStoreErr := ds.AddEvent("123", &models.Event{Type: "type", Message: "321", Time: "000", MessageOffset: "0001"})
		So(dataStoreErr, ShouldBeNil)
	})
}

func TestDataStoreGetDimensions(t *testing.T) {
	Convey("When adding a dimension, no error is returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		db.Begin()
		ds, err := NewPostgresDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare("SELECT instance FROM Instances WHERE instanceId").ExpectQuery().WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"job"}).AddRow("{}"))
		mock.ExpectPrepare("SELECT nodeName, value, nodeId").ExpectQuery().
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"nodeName", "value", "nodeId"}).
				AddRow("node1", "0", "1").AddRow("node2", "2", "2"))
		dimensions, dataStoreErr := ds.GetDimension("123")
		So(dataStoreErr, ShouldBeNil)
		So(len(dimensions), ShouldEqual, 2)
	})
}

func TestDataStoreAddNodeId(t *testing.T) {
	Convey("", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		db.Begin()
		ds, err := NewPostgresDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare("UPDATE Dimensions SET nodeId").ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"instanceId"}).AddRow("123"))
		dataStoreErr := ds.AddNodeId("123", "node1", &models.Dimension{NodeId: "123"})
		So(dataStoreErr, ShouldBeNil)
	})
}

func NewSQLMockWithSQLStatements() (sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	So(err, ShouldBeNil)
	mock.ExpectBegin()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectPrepare("INSERT INTO Jobs")
	mock.ExpectPrepare("UPDATE Jobs SET job = jsonb_set")
	mock.ExpectPrepare("INSERT INTO Instances")
	mock.ExpectPrepare("SELECT instance FROM Instances WHERE")
	mock.ExpectPrepare("UPDATE Jobs SET job = jsonb_set")
	mock.ExpectPrepare("UPDATE Instances SET instance = jsonb_set")
	mock.ExpectPrepare("INSERT INTO Dimensions")
	mock.ExpectPrepare("SELECT nodeName, value, nodeId")
	mock.ExpectPrepare("UPDATE Dimensions SET nodeId")
	return mock, db
}

//addJob := prepare("INSERT INTO Jobs(job) VALUES($1) RETURNING jobId", db) UPDATE Jobs SET job = jsonb_set
//addInstance := prepare("INSERT INTO Instances(instance) VALUES($1) RETURNING instanceId", db)
//findInstance := prepare("SELECT instance FROM Instances WHERE instanceId = $1", db)
//addS3Url := prepare("UPDATE Jobs SET job = jsonb_set(job, '{s3Files}', (SELECT (job->'s3Files')  || TO_JSONB(json_build_object('alaisName',$1::TEXT,'url',$2::TEXT)) FROM Jobs WHERE jobId = $3), true) WHERE jobId = $3 RETURNING jobId", db)
//addEvent := prepare("UPDATE Instances SET instance = jsonb_set(instance, '{events}', (SELECT (instance->'events')  || TO_JSONB(json_build_object('type', $1::TEXT, 'time', $2::TEXT, 'message', $3::TEXT, 'messageOffset', $4::TEXT)) FROM Jobs WHERE instanceid = $5), true) WHERE instanceid = $5 RETURNING instanceId", db)
//addDimension := prepare("INSERT INTO Dimensions(instanceId, nodeName, value) VALUES($1, $2, $3)", db)
//getDimensions := prepare("SELECT nodeName, value, nodeId FROM Dimensions WHERE instanceId = $1", db)
//addNodeId := prepare("UPDATE Dimensions SET nodeId = $1 WHERE instanceId = $2 AND nodeName = $3 RETURNING instanceId", db)
