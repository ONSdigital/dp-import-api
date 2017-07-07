package datastore

import (
	"database/sql"
	"encoding/json"
	"github.com/ONSdigital/dp-import-api/models"
	"time"
)

type PostgresDatastore struct {
	db            *sql.DB
	addJob        *sql.Stmt
	updateState   *sql.Stmt
	addInstance   *sql.Stmt
	findInstance  *sql.Stmt
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
	addJob := prepare("INSERT INTO Jobs(job) VALUES($1) RETURNING jobId", db)
	updateState := prepare("UPDATE Jobs SET job = jsonb_set(job, '{state}', TO_JSONB($1::TEXT)) WHERE jobId = $2 RETURNING jobId", db)
	addInstance := prepare("INSERT INTO Instances(instance) VALUES($1) RETURNING instanceId", db)
	findInstance := prepare("SELECT instance FROM Instances WHERE instanceId = $1", db)
	addS3Url := prepare("UPDATE Jobs SET job = jsonb_set(job, '{s3Files}', (SELECT (job->'s3Files')  || TO_JSONB(json_build_object('alaisName',$1::TEXT,'url',$2::TEXT)) FROM Jobs WHERE jobId = $3), true) WHERE jobId = $3 RETURNING jobId", db)
	addEvent := prepare("UPDATE Instances SET instance = jsonb_set(instance, '{events}', (SELECT (instance->'events')  || TO_JSONB(json_build_object('type', $1::TEXT, 'time', $2::TEXT, 'message', $3::TEXT, 'messageOffset', $4::TEXT)) FROM Jobs WHERE instanceid = $5), true) WHERE instanceid = $5 RETURNING instanceId", db)
	addDimension := prepare("INSERT INTO Dimensions(instanceId, nodeName, value) VALUES($1, $2, $3)", db)
	getDimensions := prepare("SELECT nodeName, value, nodeId FROM Dimensions WHERE instanceId = $1", db)
	addNodeId := prepare("UPDATE Dimensions SET nodeId = $1 WHERE instanceId = $2 AND nodeName = $3 RETURNING instanceId", db)
	return PostgresDatastore{db: db, addJob: addJob, updateState: updateState, addInstance: addInstance, findInstance: findInstance,
		addS3Url: addS3Url, addEvent: addEvent, addDimension: addDimension, getDimensions: getDimensions, addNodeId: addNodeId}, nil
}

func (ds PostgresDatastore) AddJob(newjob *models.ImportJob) (models.JobInstance, error) {
	job := models.Job{Datasets: newjob.Datasets, Recipe: newjob.Recipe, S3Files: make([]models.S3File, 0), State: "Created"}
	bytes, error := json.Marshal(job)
	if error != nil {
		return models.JobInstance{}, error
	}
	row := ds.addJob.QueryRow(bytes)
	var jobId sql.NullString
	rowError := row.Scan(&jobId)
	if rowError != nil {
		return models.JobInstance{}, rowError
	}
	instanceIds := []string{}
	for i := 0; i < len(newjob.Datasets); i++ {
		id, instanceIdErr := ds.AddInstance(newjob.Datasets[i])
		if instanceIdErr != nil {
			return models.JobInstance{}, instanceIdErr
		}
		instanceIds = append(instanceIds, id)
	}
	return models.JobInstance{JobId: jobId.String, InstanceIds: instanceIds}, nil
}

func (ds PostgresDatastore) AddS3File(instanceId string, message *models.S3File) error {
	row := ds.addS3Url.QueryRow(message.AliasName, message.Url, instanceId)
	var returnedInstanceID sql.NullString
	// Check that a instanceId is returned if not, no rows where update so return a job not found error.
	error := row.Scan(&returnedInstanceID)
	return convertError(error)
}

func (ds PostgresDatastore) UpdateJobState(jobId string, state *models.JobState) error {
	row := ds.updateState.QueryRow(jobId, state.State)
	var jobIdReturned sql.NullString
	// Check that a instanceId is returned if not, no rows where update so return a job not found error.
	dataStoreError := row.Scan(&jobIdReturned)
	if dataStoreError != nil {
		return dataStoreError
	}
	return nil
}

func (ds PostgresDatastore) AddInstance(dataset string) (string, error) {
	job := models.ImportJobState{Dataset: dataset, State: "Created", LastUpdated: time.Now().UTC().String(), S3Files: []models.S3File{}, Events: []models.Event{}}
	bytes, error := json.Marshal(job)
	if error != nil {
		return "", error
	}
	row := ds.addInstance.QueryRow(bytes)
	var instanceId sql.NullString
	rowError := row.Scan(&instanceId)
	if rowError != nil {
		return "", rowError
	}
	return instanceId.String, nil
}

func (ds PostgresDatastore) GetInstance(instanceId string) (models.ImportJobState, error) {
	row := ds.findInstance.QueryRow(instanceId)
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

func (ds PostgresDatastore) AddEvent(instanceId string, event *models.Event) error {
	row := ds.addEvent.QueryRow(event.Type, event.Time, event.Message, event.MessageOffset, instanceId)
	var returnedInstanceID sql.NullString
	// Check that a instanceId is returned if not, no rows where update so return a job not found error.
	error := row.Scan(&returnedInstanceID)
	return convertError(error)
}

func (ds PostgresDatastore) AddDimension(instanceId string, dimension *models.Dimension) error {
	_, err := ds.GetInstance(instanceId)
	if err != nil {
		return err
	}
	// Connection leak!!!!!!!!!!!!!!!!
	_, queryError := ds.addDimension.Query(instanceId, dimension.NodeName, dimension.Value)
	return queryError
}

func (ds PostgresDatastore) GetDimension(instanceId string) ([]models.Dimension, error) {
	_, err := ds.GetInstance(instanceId)
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
