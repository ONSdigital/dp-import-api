package postgres

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/log"
	pg "github.com/lib/pq"
	"github.com/satori/go.uuid"
)

var allFilterStates = []string{"created", "submitted", "completed", "error"}

// Datastore to hold SQL statements to be used to gather information or to insert Jobs and instances
type Datastore struct {
	db                        *sql.DB
	addJob                    *sql.Stmt
	getJob                    *sql.Stmt
	getJobs                   *sql.Stmt
	updateJobNoRestrictions   *sql.Stmt
	updateJobWithRestrictions *sql.Stmt
	addInstance               *sql.Stmt
	findInstance              *sql.Stmt
	getInstances              *sql.Stmt
	updateInstance            *sql.Stmt
	addFileToJob              *sql.Stmt
	addEvent                  *sql.Stmt
	addDimension              *sql.Stmt
	getDimensions             *sql.Stmt
	getDimensionValues        *sql.Stmt
	addNodeID                 *sql.Stmt
	prepareImportJob          *sql.Stmt
	incrementObservationCount *sql.Stmt
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
	updateJobNoRestrictions := prepare("UPDATE Jobs set job = job || jsonb($1::TEXT) WHERE jobId = $2 RETURNING jobId", db)
	updateJobWithRestrictions := prepare("UPDATE Jobs set job = job || jsonb($1::TEXT) WHERE jobId = $2 AND job->>'state' = 'created' RETURNING jobId", db)
	addFileToJob := prepare("UPDATE Jobs SET job = jsonb_set(job, '{files}', (SELECT (job->'files')  || TO_JSONB(json_build_object('alaisName',$1::TEXT,'url',$2::TEXT)) FROM Jobs WHERE jobId = $3), true) WHERE jobId = $3 RETURNING jobId", db)
	addInstance := prepare("INSERT INTO Instances(instanceId, jobId, instance) VALUES($1, $2, $3) RETURNING instanceId", db)
	findInstance := prepare("SELECT instance, jobId FROM Instances WHERE instanceId = $1", db)
	getInstances := prepare("SELECT instanceId, instance, jobID FROM Instances WHERE instance->>'state' = ANY ($1::TEXT[])", db)
	updateInstance := prepare("UPDATE Instances set instance = instance || jsonb($1::TEXT) WHERE instanceId = $2 RETURNING instanceId", db)
	addEvent := prepare("UPDATE Instances SET instance = jsonb_set(instance, '{events}', (SELECT (instance->'events')  || TO_JSONB(json_build_object('type', $1::TEXT, 'time', $2::TEXT, 'message', $3::TEXT, 'messageOffset', $4::TEXT)) FROM Instances WHERE instanceid = $5), true) WHERE instanceid = $5 RETURNING instanceId", db)
	addDimension := prepare("INSERT INTO Dimensions(instanceId, dimensionName, value) VALUES($1, $2, $3)", db)
	getDimensions := prepare("SELECT dimensionName, value, nodeId FROM Dimensions WHERE instanceId = $1", db)
	getDimensionValues := prepare("SELECT dimensions.value FROM dimensions WHERE instanceid = $1 AND dimensionname = $2", db)
	addNodeID := prepare("UPDATE Dimensions SET nodeId = $1 WHERE value = $2 AND instanceId = $3 AND dimensionName = $4 RETURNING instanceId", db)
	prepareImportJob := prepare("SELECT job->>'recipe', job->'files', STRING_AGG(instanceId::TEXT, ', ') FROM Jobs INNER JOIN  Instances ON (Jobs.jobId = Instances.jobId) WHERE jobs.jobId = $1 GROUP BY jobs.job", db)
	incrementObservationCount := prepare("UPDATE Instances SET instance = instance ||  jsonb(json_build_object('total_inserted_observations', (instance->>'total_inserted_observations')::int + $1)) WHERE instanceId = $2", db)

	return Datastore{db: db, addJob: addJob, getJob: getJob, getJobs: getJobs, updateJobNoRestrictions: updateJobNoRestrictions, updateJobWithRestrictions: updateJobWithRestrictions,
		addInstance: addInstance, updateInstance: updateInstance, findInstance: findInstance, getInstances: getInstances, addFileToJob: addFileToJob, addEvent: addEvent,
		addDimension: addDimension, getDimensions: getDimensions, getDimensionValues: getDimensionValues, addNodeID: addNodeID,
		prepareImportJob: prepareImportJob, incrementObservationCount: incrementObservationCount}, nil
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

	newjob.Instances = []models.IDLink{models.IDLink{ID: id, Link: url}}
	return *newjob, nil
}

// GetJobs returns a list of import jobs
func (ds Datastore) GetJobs(host string, filter []string) ([]models.Job, error) {
	if len(filter) == 0 {
		filter = allFilterStates
	}
	rows, err := ds.getJobs.Query(pg.Array(filter))
	if err != nil {
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
		job.JobID = jobID.String
		job.Instances = []models.IDLink{models.IDLink{ID: instanceID.String, Link: buildInstanceURL(host, instanceID.String)}}
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
		return models.Job{}, convertError(err)
	}
	var job models.Job
	err = json.Unmarshal([]byte(jobInfo.String), &job)
	if err != nil {
		return models.Job{}, err
	}
	job.JobID = jobID

	job.Instances = []models.IDLink{models.IDLink{ID: instanceID.String, Link: buildInstanceURL(host, instanceID.String)}}

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
func (ds Datastore) UpdateJobState(jobID string, job *models.Job, withOutRestrictions bool) error {
	// Check that the job exists
	_, err := ds.GetJob("", jobID)
	if err != nil {
		return err
	}
	json, err := json.Marshal(job)
	if err != nil {
		return err
	}
	var updateStmt *sql.Stmt
	if withOutRestrictions {
		updateStmt = ds.updateJobNoRestrictions
	} else {
		updateStmt = ds.updateJobWithRestrictions
	}

	row := updateStmt.QueryRow(string(json), jobID)
	var jobIDReturned sql.NullString
	err = row.Scan(&jobIDReturned)
	// If no rows where updated but the job exists the request didn't have the right to update job's state.
	if err == sql.ErrNoRows {
		return api_errors.ForbiddenOperation
	}
	return err
}

// AddInstance which relates to a job
func (ds Datastore) AddInstance(tx *sql.Tx, jobID string) (string, error) {
	job := models.Instance{State: "created", LastUpdated: time.Now().UTC().String(),
		Events: &[]models.Event{}, TotalObservations: new(int), InsertedObservations: new(int)}
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
func (ds Datastore) GetInstance(host, instanceID string) (models.Instance, error) {
	row := ds.findInstance.QueryRow(instanceID)
	var instanceJSON, jobID sql.NullString
	err := row.Scan(&instanceJSON, &jobID)
	if err != nil {
		return models.Instance{}, convertError(err)
	}
	var instance models.Instance
	err = json.Unmarshal([]byte(instanceJSON.String), &instance)
	if err != nil {
		return models.Instance{}, err
	}
	instance.InstanceID = instanceID
	instance.Job = models.IDLink{ID: jobID.String, Link: buildJobURL(host, jobID.String)}
	return instance, nil
}

// GetInstances from postgres
func (ds Datastore) GetInstances(host string, filter []string) ([]models.Instance, error) {
	if len(filter) == 0 {
		filter = allFilterStates
	}
	var instances []models.Instance
	rows, err := ds.getInstances.Query(pg.Array(filter))
	if err != nil {
		return []models.Instance{}, err
	}
	for rows.Next() {
		var instanceID, instanceJSON, jobID sql.NullString
		err = rows.Scan(&instanceID, &instanceJSON, &jobID)
		if err != nil {
			return []models.Instance{}, err
		}
		var instance models.Instance
		err = json.Unmarshal([]byte(instanceJSON.String), &instance)
		if err != nil {
			return []models.Instance{}, err
		}
		instance.InstanceID = instanceID.String
		instance.Job = models.IDLink{ID: jobID.String, Link: buildJobURL(host, jobID.String)}
		instances = append(instances, instance)
	}
	return instances, nil
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
	return convertError(row.Scan(&instanceIDReturned))
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
	_, err := ds.GetInstance("", instanceID)
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
	_, err := ds.GetInstance("", instanceID)
	if err != nil {
		return []models.Dimension{}, err
	}
	rows, err := ds.getDimensions.Query(instanceID)
	if err != nil {
		return []models.Dimension{}, convertError(err)
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

// GetDimensionValues from a store dimension in postgres, each value returned is unique
func (ds Datastore) GetDimensionValues(instanceID, dimensionName string) (models.UniqueDimensionValues, error) {
	values := []string{}
	rows, err := ds.getDimensionValues.Query(instanceID, dimensionName)
	if err != nil {
		return models.UniqueDimensionValues{}, err
	}
	for rows.Next() {
		var value sql.NullString
		err := rows.Scan(&value)
		if err != nil {
			return models.UniqueDimensionValues{}, err
		}
		values = append(values, value.String)
	}

	if len(values) == 0 {
		return models.UniqueDimensionValues{}, api_errors.DimensionNameNotFoundError
	}

	return models.UniqueDimensionValues{Name: dimensionName, Values: values}, nil
}

// AddNodeID for a dimension
func (ds Datastore) AddNodeID(instanceID string, dimension *models.Dimension) error {
	row := ds.addNodeID.QueryRow(dimension.NodeID, dimension.Value, instanceID, dimension.Name)
	var returnedInstanceID sql.NullString
	return convertError(row.Scan(&returnedInstanceID))
}

// PrepareImportJob to send to data baker
func (ds Datastore) PrepareImportJob(jobID string) (*models.ImportData, error) {
	tx, err := ds.db.Begin()
	if err != nil {
		return nil, err
	}
	row := tx.Stmt(ds.prepareImportJob).QueryRow(jobID)
	var recipe, filesAsJSON, instanceIds sql.NullString
	err = row.Scan(&recipe, &filesAsJSON, &instanceIds)
	if err != nil {
		return nil, err
	}
	var files []models.UploadedFile
	err = json.Unmarshal([]byte(filesAsJSON.String), &files)
	if err != nil {
		return nil, err
	}
	instances := strings.Split(instanceIds.String, ",")
	for _, instance := range instances {
		log.Info(instance, log.Data{})
		json, err := json.Marshal(models.Instance{State: "submitted"})
		if err != nil {
			return nil, tx.Rollback()
		}
		_, err = tx.Stmt(ds.updateInstance).Exec(string(json), instance)
		if err != nil {
			return nil, tx.Rollback()
		}
	}
	return &models.ImportData{JobID: jobID,
		Recipe:        recipe.String,
		UploadedFiles: files,
		InstanceIDs:   strings.Split(instanceIds.String, ",")}, tx.Commit()
}

func (ds Datastore) UpdateObservationCount(instanceID string, count int) error {
	results, err := ds.incrementObservationCount.Exec(count, instanceID)
	if err != nil {
		return err
	}
	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return api_errors.JobNotFoundError
	}
	return nil
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

func buildInstanceURL(host, id string) string {
	return host + "/instances/" + id
}

func buildJobURL(host, id string) string {
	return host + "/jobs/" + id
}
