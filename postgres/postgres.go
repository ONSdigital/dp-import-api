package postgres

import (
	"database/sql"
	"encoding/json"
	"github.com/ONSdigital/dp-import-api/errors"
	"github.com/ONSdigital/dp-import-api/models"
	"strings"
	"time"
)

// Datastore - A structure to hold SQL statements to be used to gather information or insert about Jobs and instances
type Datastore struct {
	db                   *sql.DB
	addJob               *sql.Stmt
	updateJob            *sql.Stmt
	addInstance          *sql.Stmt
	findInstance         *sql.Stmt
	addFileToJob         *sql.Stmt
	addEvent             *sql.Stmt
	addDimension         *sql.Stmt
	getDimensions        *sql.Stmt
	addNodeID            *sql.Stmt
	createPublishMessage *sql.Stmt
}

func prepare(sql string, db *sql.DB) *sql.Stmt {
	statement, err := db.Prepare(sql)
	if err != nil {
		panic(err)
	}
	return statement
}

// NewDatastore - Create a postgres datastore. This is used to store and find information about jobs and instances.
func NewDatastore(db *sql.DB) (Datastore, error) {
	addJob := prepare("INSERT INTO Jobs(job) VALUES($1) RETURNING jobId", db)
	updateJob := prepare("UPDATE Jobs set job = job || jsonb($1::TEXT) WHERE jobId = $2 RETURNING jobId", db)
	addFileToJob := prepare("UPDATE Jobs SET job = jsonb_set(job, '{files}', (SELECT (job->'files')  || TO_JSONB(json_build_object('alaisName',$1::TEXT,'url',$2::TEXT)) FROM Jobs WHERE jobId = $3), true) WHERE jobId = $3 RETURNING jobId", db)
	addInstance := prepare("INSERT INTO Instances(jobId, instance) VALUES($1, $2) RETURNING instanceId", db)
	findInstance := prepare("SELECT instance FROM Instances WHERE instanceId = $1", db)
	addEvent := prepare("UPDATE Instances SET instance = jsonb_set(instance, '{events}', (SELECT (instance->'events')  || TO_JSONB(json_build_object('type', $1::TEXT, 'time', $2::TEXT, 'message', $3::TEXT, 'messageOffset', $4::TEXT)) FROM Instances WHERE instanceid = $5), true) WHERE instanceid = $5 RETURNING instanceId", db)
	addDimension := prepare("INSERT INTO Dimensions(instanceId, nodeName, value) VALUES($1, $2, $3)", db)
	getDimensions := prepare("SELECT nodeName, value, nodeId FROM Dimensions WHERE instanceId = $1", db)
	addNodeID := prepare("UPDATE Dimensions SET nodeId = $1 WHERE instanceId = $2 AND nodeName = $3 RETURNING instanceId", db)
	createPublishMessage := prepare("SELECT job->>'recipe', job->'files', STRING_AGG(instanceId::TEXT, ', ') FROM Jobs INNER JOIN  Instances ON (Jobs.jobId = Instances.jobId) WHERE jobs.jobId = $1 GROUP BY jobs.job", db)
	return Datastore{db: db, addJob: addJob, updateJob: updateJob, addInstance: addInstance,
		findInstance: findInstance, addFileToJob: addFileToJob, addEvent: addEvent, addDimension: addDimension,
		getDimensions: getDimensions, addNodeID: addNodeID, createPublishMessage: createPublishMessage}, nil
}

// AddJob - Add a job to be stored in postgres.
func (ds Datastore) AddJob(host string, newjob *models.Job) (models.Job, error) {
	bytes, error := json.Marshal(newjob)
	if error != nil {
		return models.Job{}, error
	}
	tx, _ := ds.db.Begin()
	row := tx.Stmt(ds.addJob).QueryRow(bytes)
	var jobID sql.NullString
	rowError := row.Scan(&jobID)
	if rowError != nil {
		return models.Job{}, rowError
	}

	id, instanceIDErr := ds.AddInstance(tx, jobID.String)
	if instanceIDErr != nil {
		if txError := tx.Rollback(); txError != nil {
			return models.Job{}, txError
		}
		return models.Job{}, instanceIDErr
	}
	if txError := tx.Commit(); txError != nil {
		return models.Job{}, txError
	}
	url := host + "/instances/" + id
	newjob.JobID = jobID.String
	newjob.Links.InstanceIDs = []string{url}
	return *newjob, nil
}

// AddUploadedFile -  Add an uploaded file to a job.
func (ds Datastore) AddUploadedFile(instanceID string, message *models.UploadedFile) error {
	row := ds.addFileToJob.QueryRow(message.AliasName, message.URL, instanceID)
	var returnedInstanceID sql.NullString
	// Check that a instanceID is returned if not, no rows where update so return a job not found error
	error := row.Scan(&returnedInstanceID)
	return convertError(error)
}

// UpdateJobState - Update the state of a job.
func (ds Datastore) UpdateJobState(jobID string, job *models.Job) error {
	json, jsonErr := json.Marshal(job)
	if jsonErr != nil {
		return jsonErr
	}
	row := ds.updateJob.QueryRow(string(json), jobID)
	var jobIDReturned sql.NullString
	// Check that a instanceId is returned if not, no rows where update so return a job not found error
	dataStoreError := row.Scan(&jobIDReturned)
	return dataStoreError
}

// AddInstance - Add an instance and relate it to a job.
func (ds Datastore) AddInstance(tx *sql.Tx, jobID string) (string, error) {
	job := models.Instance{State: "Created", LastUpdated: time.Now().UTC().String(), Events: []models.Event{}}
	bytes, error := json.Marshal(job)
	if error != nil {
		return "", error
	}
	row := tx.Stmt(ds.addInstance).QueryRow(jobID, bytes)
	var instanceID sql.NullString
	rowError := row.Scan(&instanceID)
	if rowError != nil {
		return "", rowError
	}
	return instanceID.String, nil
}

// GetInstance - Get an instance from postgres.
func (ds Datastore) GetInstance(instanceID string) (models.Instance, error) {
	row := ds.findInstance.QueryRow(instanceID)
	var job sql.NullString
	rowError := row.Scan(&job)
	if rowError != nil {
		return models.Instance{}, convertError(rowError)
	}
	var importJob models.Instance
	error := json.Unmarshal([]byte(job.String), &importJob)
	if error != nil {
		return models.Instance{}, error
	}
	importJob.InstanceID = instanceID
	return importJob, nil
}

// AddEvent - Add an event into an instance.
func (ds Datastore) AddEvent(instanceID string, event *models.Event) error {
	row := ds.addEvent.QueryRow(event.Type, event.Time, event.Message, event.MessageOffset, instanceID)
	var returnedInstanceID sql.NullString
	// Check that a instanceID is returned if not, no rows where update so return a job not found error
	error := row.Scan(&returnedInstanceID)
	return convertError(error)
}

// AddDimension - Add a dimension to cache in postgres
func (ds Datastore) AddDimension(instanceID string, dimension *models.Dimension) error {
	// Check that an instance exists else return an error
	_, err := ds.GetInstance(instanceID)
	if err != nil {
		return err
	}
	res, queryError := ds.addDimension.Query(instanceID, dimension.Name, dimension.Value)
	if queryError != nil {
		return queryError
	}
	return res.Close()
}

// GetDimension - Get all dimensions related to an instanceID
func (ds Datastore) GetDimension(instanceID string) ([]models.Dimension, error) {
	_, err := ds.GetInstance(instanceID)
	if err != nil {
		return []models.Dimension{}, err
	}
	rows, err := ds.getDimensions.Query(instanceID)
	if err != nil {
		return []models.Dimension{}, err
	}
	dimensions := []models.Dimension{}
	for rows.Next() {
		var nodeName, value, nodeID sql.NullString
		err := rows.Scan(&nodeName, &value, &nodeID)
		if err != nil {
			return []models.Dimension{}, err
		}
		dimensions = append(dimensions, models.Dimension{Name: nodeName.String, NodeID: nodeID.String, Value: value.String})
	}
	return dimensions, nil
}

// AddNodeID - Add nodeID for a dimension.
func (ds Datastore) AddNodeID(instanceID, nodeID string, message *models.Dimension) error {
	row := ds.addNodeID.QueryRow(message.NodeID, instanceID, nodeID)
	var returnedInstanceID sql.NullString
	error := row.Scan(&returnedInstanceID)
	return convertError(error)
}

// BuildImportDataMessage - Build a publish message to send to data baker
func (ds Datastore) BuildImportDataMessage(jobID string) (*models.ImportData, error) {
	row := ds.createPublishMessage.QueryRow(jobID)
	var recipe, filesAsJSON, instancIds sql.NullString
	error := row.Scan(&recipe, &filesAsJSON, &instancIds)
	if error != nil {
		return nil, error
	}
	var files []models.UploadedFile
	err := json.Unmarshal([]byte(filesAsJSON.String), &files)
	if err != nil {
		return nil, err
	}
	return &models.ImportData{JobId: jobID,
		Recipe:        recipe.String,
		UploadedFiles: files,
		InstanceIds:   strings.Split(instancIds.String, ",")}, nil
}

func convertError(err error) error {
	switch {
	case err == sql.ErrNoRows:
		return errors.JobNotFoundError
	case err != nil:
		return err
	}
	return nil
}
