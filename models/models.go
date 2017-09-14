package models

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"time"
)

// JobResults for list of Job items
type JobResults struct {
	Items []Job `json:"items"`
}

// Job for importing datasets
type Job struct {
	JobID         string          `bson:"job_id,omitempty"           json:"job_id,omitempty"`
	Recipe        string          `bson:"recipe,omitempty"           json:"recipe,omitempty"`
	State         string          `bson:"state,omitempty"            json:"state,omitempty"`
	UploadedFiles *[]UploadedFile `bson:"files,omitempty"            json:"files,omitempty"`
	Links         LinksMap        `bson:"links,omitempty"            json:"links,omitempty"`
	LastUpdated   time.Time       `bson:"last_updated,omitempty"     json:"last_updated,omitempty"`
}

type LinksMap struct {
	Instances []IDLink `bson:"instances,omitempty" json:"instances,omitempty"`
}

// Validate the content of a job
func (job *Job) Validate() error {
	if job.Recipe == "" {
		return errors.New("Missing properties to create importqueue job struct")
	}
	if job.State == "" {
		job.State = "created"
	}
	if job.UploadedFiles == nil {
		job.UploadedFiles = &[]UploadedFile{}
	}
	return nil
}

// Event which has happened to an instance
type Event struct {
	Type          string `json:"type"`
	Time          string `json:"time"`
	Message       string `json:"message"`
	MessageOffset string `json:"messageOffset"`
}

// Instance which presents a single dataset being imported
type Instance struct {
	InstanceID           string         `json:"id,omitempty"`
	Links                *InstanceLinks `json:"links,omitempty"`
	State                string         `json:"state,omitempty"`
	Events               []Event        `json:"events,omitempty"`
	TotalObservations    int            `json:"total_observations,omitempty"`
	InsertedObservations int            `json:"total_inserted_observations,omitempty"`
	Headers              []string       `json:"headers,omitempty"`
	LastUpdated          string         `json:"last_updated,omitempty"`
}

type InstanceLinks struct {
	Job IDLink `json:"job,omitempty"`
}

// UploadedFile used for a file which has been uploaded to a bucket
type UploadedFile struct {
	AliasName string `bson:"alias_name" json:"alias_name" avro:"alias-name"`
	URL       string `bson:"url"        json:"url"        avro:"url"`
}

// Validate the content of the structure
func (s UploadedFile) Validate() error {
	if s.URL == "" || s.AliasName == "" {
		return errors.New("Invalid s3 file structure")
	}
	return nil
}

// ImportData used to create a message to data baker or direct to the dimension-extractor
type ImportData struct {
	JobID         string
	Recipe        string          `json:"recipe,omitempty"`
	UploadedFiles *[]UploadedFile `json:"files,omitempty"`
	InstanceIDs   []string
}

// DataBakerEvent used to trigger the databaker process
type DataBakerEvent struct {
	JobID string `avro:"job_id"`
}

// IDLink holds the id and a link to the resource
type IDLink struct {
	ID   string `json:"id"`
	HRef string `json:"href"`
}

// CreateJob from a json message
func CreateJob(reader io.Reader) (*Job, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}
	var job Job
	err = json.Unmarshal(bytes, &job)
	if err != nil {
		return nil, errors.New("Failed to parse json body")
	}
	return &job, nil
}

// CreateUploadedFile from a json message
func CreateUploadedFile(reader io.Reader) (*UploadedFile, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}
	var message UploadedFile
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		return nil, errors.New("Failed to parse json body")
	}
	return &message, message.Validate()
}

// CreateInstance from a job ID
func CreateInstance(jobID, jobURL string) *Instance {
	return &Instance{
		Links: &InstanceLinks{
			Job: IDLink{ID: jobID, HRef: jobURL},
		}}
}
