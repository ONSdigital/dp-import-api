package postgres

import (
	"database/sql"
	"encoding/json"
	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/models"
	"strings"
	"time"
	"github.com/satori/go.uuid"
    pg "github.com/lib/pq"
)

var allFilterStates = []string{"created", "submitted", "completed", "error"}

// Datastore to hold SQL statements to be used to gather information or to insert Jobs and instances
type Datastore struct {
	db                   *sql.DB
	addJob               *sql.Stmt
	getJob               *sql.Stmt
	getJobs              *sql.Stmt
	updateJob            *sql.Stmt
	addInstance          *sql.Stmt
	findInstance         *sql.Stmt
	updateInstance       *sql.Stmt
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

// NewDatastore used to store jobs and instances in postgres
func NewDatastore(db *sql.DB) (Datastore, error) {
	addJob := prepare("INSERT INTO Jobs(jobid,job) VALUES($1, $2) RETURNING jobId", db)
	getJob := prepare("SELECT instanceId, job FROM Jobs INNER JOIN  Instances ON (Jobs.jobId = Instances.jobId) WHERE Jobs.jobId = $1 ", db)
	getJobs := prepare("SELECT Jobs.jobId, instanceId, job FROM Jobs INNER JOIN  Instances ON (Jobs.jobId = Instances.jobId) WHERE Jobs.job->>'state' = ANY ($1::TEXT[])", db)
	updateJob := prepare("UPDATE Jobs set job = job || jsonb($1::TEXT) WHERE jobId = $2 RETURNING jobId", db)
	addFileToJob := prepare("UPDATE Jobs SET job = jsonb_set(job, '{files}', (SELECT (job->'files')  || TO_JSONB(json_build_object('alaisName',$1::TEXT,'url',$2::TEXT)) FROM Jobs WHERE jobId = $3), true) WHERE jobId = $3 RETURNING jobId", db)
	addInstance := prepare("INSERT INTO Instances(instanceId, jobId, instance) VALUES($1, $2, $3) RETURNING instanceId", db)
	findInstance := prepare("SELECT instance FROM Instances WHERE instanceId = $1", db)
	updateInstance := prepare("UPDATE Instances set instance = instance || jsonb($1::TEXT) WHERE instanceId = $2 RETURNING instanceId", db)
	addEvent := prepare("UPDATE Instances SET instance = jsonb_set(instance, '{events}', (SELECT (instance->'events')  || TO_JSONB(json_build_object('type', $1::TEXT, 'time', $2::TEXT, 'message', $3::TEXT, 'messageOffset', $4::TEXT)) FROM Instances WHERE instanceid = $5), true) WHERE instanceid = $5 RETURNING instanceId", db)
	addDimension := prepare("INSERT INTO Dimensions(instanceId, dimensionName, value) VALUES($1, $2, $3)", db)
	getDimensions := prepare("SELECT dimensionName, value, nodeId FROM Dimensions WHERE instanceId = $1", db)
	addNodeID := prepare("UPDATE Dimensions SET nodeId = $1 WHERE instanceId = $2 AND dimensionName = $3 RETURNING instanceId", db)
	createPublishMessage := prepare("SELECT job->>'recipe', job->'files', STRING_AGG(instanceId::TEXT, ', ') FROM Jobs INNER JOIN  Instances ON (Jobs.jobId = Instances.jobId) WHERE jobs.jobId = $1 GROUP BY jobs.job", db)
	return Datastore{db: db, addJob: addJob, getJob: getJob, getJobs: getJobs, updateJob: updateJob, addInstance: addInstance, updateInstance: updateInstance,
		findInstance: findInstance, addFileToJob: addFileToJob, addEvent: addEvent, addDimension: addDimension,
		getDimensions: getDimensions, addNodeID: addNodeID, createPublishMessage: createPublishMessage}, nil
}

// AddJob store a job in postgres
func (ds Datastore) AddJob(host string, newjob *models.Job) (models.Job, error) {
	bytes, err := json.Marshal(newjob)
	if err != nil {
		return models.Job{}, err
	}
	tx, err := ds.db.Begin()
	if err != nil {
		return models.Job{}, err
	}
	uuid := uuid.NewV4().String()
	row := tx.Stmt(ds.addJob).QueryRow(uuid, bytes)
	var jobID sql.NullString
	err = row.Scan(&jobID)
	if err != nil {
		return models.Job{}, err
	}

	id, err := ds.AddInstance(tx, jobID.String)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return models.Job{}, err
		}
		return models.Job{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Job{}, err
	}
	url := host + "/instances/" + id
	newjob.JobID = jobID.String
	newjob.Links.InstanceIDs = []string{url}
	return *newjob, nil
}

// GetJobs returns a list of import jobs
func (ds Datastore) GetJobs(host string, filter []string) ([]models.Job, error) {
	if len(filter) == 0 {
		filter = allFilterStates
	}
	rows, err := ds.getJobs.Query(pg.Array(filter))
	if err !=nil {
		return []models.Job{}, err
	}

	jobs := []models.Job{}
	for rows.Next() {
		var jobID, instanceID, jobInfo sql.NullString
		err = rows.Scan(&jobID, &instanceID, &jobInfo)
		if err != nil {
			return []models.Job{}, err
		}
		var job models.Job
		err = json.Unmarshal([]byte(jobInfo.String), &job)
		if err != nil {
			return []models.Job{}, err
		}
		url := host + "/instances/" + instanceID.String
		job.JobID = jobID.String
		job.Links.InstanceIDs = []string{url}
		jobs = append(jobs, job)

	}
	return jobs, nil
}

// GetJob returns a single job from a jobID
func (ds Datastore) GetJob(host string, jobID string) (models.Job, error) {
	row := ds.getJob.QueryRow(jobID)
	var instanceID, jobInfo sql.NullString
	err := row.Scan(&instanceID, &jobInfo)
	if err != nil {
		return models.Job{}, err
	}
	var job models.Job
	err = json.Unmarshal([]byte(jobInfo.String), &job)
	if err != nil {
		return models.Job{}, err
	}
	url := host + "/instances/" + instanceID.String
	job.JobID = jobID
	job.Links.InstanceIDs = []string{url}
	return job, nil
}

// AddUploadedFile to a import job
func (ds Datastore) AddUploadedFile(instanceID string, message *models.UploadedFile) error {
	row := ds.addFileToJob.QueryRow(message.AliasName, message.URL, instanceID)
	var returnedInstanceID sql.NullString
	// Check that a instanceID is returned if not, no rows where update so return a job not found error
	return convertError(row.Scan(&returnedInstanceID))
}

// UpdateJobState configure the jobs state
func (ds Datastore) UpdateJobState(jobID string, job *models.Job) error {
	json, err := json.Marshal(job)
	if err != nil {
		return err
	}
	row := ds.updateJob.QueryRow(string(json), jobID)
	var jobIDReturned sql.NullString
	// Check that a instanceId is returned if not, no rows where update so return a job not found error
	return convertError(row.Scan(&jobIDReturned))
}

// AddInstance which relates to a job
func (ds Datastore) AddInstance(tx *sql.Tx, jobID string) (string, error) {
	job := models.Instance{State: "created", LastUpdated: time.Now().UTC().String(), Events: &[]models.Event{}}
	bytes, err := json.Marshal(job)
	if err != nil {
		return "", err
	}
	uuid := uuid.NewV4().String()
	row := tx.Stmt(ds.addInstance).QueryRow(uuid, jobID, bytes)
	var instanceID sql.NullString
	err = row.Scan(&instanceID)
	if err != nil {
		return "", err
	}
	return instanceID.String, nil
}

// GetInstance from postgres
func (ds Datastore) GetInstance(instanceID string) (models.Instance, error) {
	row := ds.findInstance.QueryRow(instanceID)
	var job sql.NullString
	err := row.Scan(&job)
	if err != nil {
		return models.Instance{}, convertError(err)
	}
	var importJob models.Instance
	err = json.Unmarshal([]byte(job.String), &importJob)
	if err != nil {
		return models.Instance{}, err
	}
	importJob.InstanceID = instanceID
	return importJob, nil
}

// UpdateInstance in postgres
func (ds Datastore) UpdateInstance(instanceID string, instance *models.Instance) error {
	json, err := json.Marshal(instance)
	if err != nil {
		return err
	}
	row := ds.updateInstance.QueryRow(string(json), instanceID)
	var instanceIDReturned sql.NullString
	// Check that a instanceId is returned if not, no rows where update so return a job not found error
	return  convertError(row.Scan(&instanceIDReturned))
}

// AddEvent into an instance
func (ds Datastore) AddEvent(instanceID string, event *models.Event) error {
	row := ds.addEvent.QueryRow(event.Type, event.Time, event.Message, event.MessageOffset, instanceID)
	var returnedInstanceID sql.NullString
	// Check that a instanceID is returned if not, no rows where update so return a job not found error
	return convertError(row.Scan(&returnedInstanceID))
}

// AddDimension to cache in postgres
func (ds Datastore) AddDimension(instanceID string, dimension *models.Dimension) error {
	// Check that an instance exists else return an error
	_, err := ds.GetInstance(instanceID)
	if err != nil {
		return err
	}
	res, err := ds.addDimension.Query(instanceID, dimension.Name, dimension.Value)
	if err != nil {
		return err
	}
	return res.Close()
}

// GetDimensions related to an instanceID
func (ds Datastore) GetDimensions(instanceID string) ([]models.Dimension, error) {
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

// AddNodeID for a dimension
func (ds Datastore) AddNodeID(instanceID, nodeID string, message *models.Dimension) error {
	row := ds.addNodeID.QueryRow(message.NodeID, instanceID, nodeID)
	var returnedInstanceID sql.NullString
	return convertError(row.Scan(&returnedInstanceID))
}

// BuildImportDataMessage to send to data baker
func (ds Datastore) BuildImportDataMessage(jobID string) (*models.ImportData, error) {
	row := ds.createPublishMessage.QueryRow(jobID)
	var recipe, filesAsJSON, instancIds sql.NullString
	err := row.Scan(&recipe, &filesAsJSON, &instancIds)
	if err != nil {
		return nil, err
	}
	var files []models.UploadedFile
	err = json.Unmarshal([]byte(filesAsJSON.String), &files)
	if err != nil {
		return nil, err
	}
	return &models.ImportData{JobID: jobID,
		Recipe:        recipe.String,
		UploadedFiles: files,
		InstanceIDs:   strings.Split(instancIds.String, ",")}, nil
}

func convertError(err error) error {
	switch {
	case err == sql.ErrNoRows:
		return api_errors.JobNotFoundError
	case err != nil:
		return err
	}
	return nil
}
