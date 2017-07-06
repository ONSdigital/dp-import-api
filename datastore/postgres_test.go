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

func TestDataStoreAddNewImportJobReturnsInstanceId(t *testing.T) {
	Convey("When creating a new import job, an instanceId is returned", t, func() {
		expectId := "000001"
		mock, db := NewSQLMockWithSQLStatements()
		db.Begin()
		ds, err := NewPostgresDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare("INSERT INTO Jobs").ExpectQuery().
			WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"instanceId"}).
			AddRow(expectId))
		jobId, err := ds.AddImportJob(&models.ImportJob{Dataset: "123"})
		So(err, ShouldBeNil)
		So(jobId.InstanceId, ShouldEqual, expectId)
	})
}

func TestDataStoreGetImportJob(t *testing.T) {
	Convey("When an instanceId is provided, the import job state is returned", t, func() {
		jsonContent := "{ \"dataset\":\"123\"  }"
		mock, db := NewSQLMockWithSQLStatements()
		db.Begin()
		ds, err := NewPostgresDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectPrepare("SELECT job FROM Jobs WHERE").ExpectQuery().
			WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"job"}).
			AddRow(jsonContent))
		state, err := ds.GetImportJob("any")
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
		mock.ExpectPrepare("UPDATE Jobs SET job = jsonb_set").ExpectQuery().
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
		mock.ExpectPrepare("SELECT job FROM Jobs WHERE").ExpectQuery().WithArgs(sqlmock.AnyArg()).
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
	mock.ExpectPrepare("SELECT job FROM Jobs WHERE")
	mock.ExpectPrepare("UPDATE Jobs SET job")
	mock.ExpectPrepare("UPDATE Jobs SET job = jsonb_set")
	mock.ExpectPrepare("INSERT INTO Dimensions")
	mock.ExpectPrepare("SELECT nodeName, value, nodeId")
	mock.ExpectPrepare("UPDATE Dimensions SET nodeId")
	return mock, db
}
