package datastore

import (
	"database/sql"
	"encoding/json"
	"github.com/ONSdigital/dp-import-api/models"
	"time"
)

type PostgresDatastore struct {
	db            *sql.DB
	newImportJob  *sql.Stmt
	findImportJob *sql.Stmt
	addS3Url      *sql.Stmt
	addEvent      *sql.Stmt
	addDimension  *sql.Stmt
	getDimensions *sql.Stmt
	addNodeId     *sql.Stmt
}

func prepare(sql string, db *sql.DB) *sql.Stmt {
	statement, err := db.Prepare(sql)
	if err != nil {
		panic(err)
	}
	return statement
}

func NewPostgresDatastore(db *sql.DB) (PostgresDatastore, error) {
	newImportJob := prepare("INSERT INTO Jobs(job) VALUES($1) RETURNING instanceId", db)
	findImportJob := prepare("SELECT job FROM Jobs WHERE instanceId = $1", db)
	addS3Url := prepare("UPDATE Jobs SET job = jsonb_set(job, '{s3Files}', (SELECT (job->'s3Files')  || TO_JSONB(json_build_object('alaisName',$1::TEXT,'url',$2::TEXT)) FROM Jobs WHERE instanceid = $3), true) WHERE instanceid = $3 RETURNING instanceId", db)
	addEvent := prepare("UPDATE Jobs SET job = jsonb_set(job, '{events}', (SELECT (job->'events')  || TO_JSONB(json_build_object('type', $1::TEXT, 'time', $2::TEXT, 'message', $3::TEXT, 'messageOffset', $4::TEXT)) FROM Jobs WHERE instanceid = $5), true) WHERE instanceid = $5 RETURNING instanceId", db)
	addDimension := prepare("INSERT INTO Dimensions(instanceId, nodeName, value) VALUES($1, $2, $3)", db)
	getDimensions := prepare("SELECT nodeName, value, nodeId FROM Dimensions WHERE instanceId = $1", db)
	addNodeId := prepare("UPDATE Dimensions SET nodeId = $1 WHERE instanceId = $2 AND nodeName = $3 RETURNING instanceId", db)
	return PostgresDatastore{db: db, newImportJob: newImportJob, findImportJob: findImportJob,
		addS3Url: addS3Url, addEvent: addEvent, addDimension: addDimension, getDimensions: getDimensions, addNodeId: addNodeId}, nil
}

func (ds PostgresDatastore) AddImportJob(newJob *models.ImportJob) (models.JobInstance, error) {
	job := models.ImportJobState{Dataset: newJob.Dataset, State: "Created", LastUpdated: time.Now().UTC().String(), S3Files: []models.S3File{}, Events: []models.Event{}}
	bytes, error := json.Marshal(job)
	if error != nil {
		return models.JobInstance{}, error
	}
	row := ds.newImportJob.QueryRow(bytes)
	var instanceId sql.NullString
	rowError := row.Scan(&instanceId)
	if rowError != nil {
		return models.JobInstance{}, rowError
	}
	return models.JobInstance{InstanceId: instanceId.String}, nil
}

func (ds PostgresDatastore) GetImportJob(instanceId string) (models.ImportJobState, error) {
	row := ds.findImportJob.QueryRow(instanceId)
	var job sql.NullString
	rowError := row.Scan(&job)
	if rowError != nil {
		return models.ImportJobState{}, convertError(rowError)
	}
	var importJob models.ImportJobState
	error := json.Unmarshal([]byte(job.String), &importJob)
	if error != nil {
		return models.ImportJobState{}, error
	}
	importJob.InstanceId = instanceId
	return importJob, nil
}

func (ds PostgresDatastore) AddS3File(instanceId string, message *models.S3File) error {
	row := ds.addS3Url.QueryRow(message.AliasName, message.S3Url, instanceId)
	var returnedInstanceID sql.NullString
	// Check that a instanceId is returned if not, no rows where update so return a job not found error.
	error := row.Scan(&returnedInstanceID)
	return convertError(error)
}

func (ds PostgresDatastore) AddEvent(instanceId string, event *models.Event) error {
	row := ds.addEvent.QueryRow(event.Type, event.Time, event.Message, event.MessageOffset, instanceId)
	var returnedInstanceID sql.NullString
	// Check that a instanceId is returned if not, no rows where update so return a job not found error.
	error := row.Scan(&returnedInstanceID)
	return convertError(error)
}

func (ds PostgresDatastore) AddDimension(instanceId string, dimension *models.Dimension) error {
	_, err := ds.GetImportJob(instanceId)
	if err != nil {
		return err
	}
	_, queryError := ds.addDimension.Query(instanceId, dimension.NodeName, dimension.Value)
	return queryError
}

func (ds PostgresDatastore) GetDimension(instanceId string) ([]models.Dimension, error) {
	_, err := ds.GetImportJob(instanceId)
	if err != nil {
		return []models.Dimension{}, err
	}
	rows, err := ds.getDimensions.Query(instanceId)
	if err != nil {
		return []models.Dimension{}, err
	}
	dimensions := []models.Dimension{}
	for rows.Next() {
		var nodeName, value, nodeId sql.NullString
		err := rows.Scan(&nodeName, &value, &nodeId)
		if err != nil {
			return []models.Dimension{}, err
		}
		dimensions = append(dimensions, models.Dimension{NodeName: nodeName.String, NodeId: nodeId.String, Value: value.String})
	}
	return dimensions, nil
}

func (ds PostgresDatastore) AddNodeId(instanceId, nodeId string, message *models.Dimension) error {
	row := ds.addNodeId.QueryRow(message.NodeId, instanceId, nodeId)
	var returnedInstanceId sql.NullString
	error := row.Scan(&returnedInstanceId)
	return convertError(error)
}

func convertError(err error) error {
	switch {
	case err == sql.ErrNoRows:
		return JobNotFoundError
	case err != nil:
		return err
	}
	return nil
}
