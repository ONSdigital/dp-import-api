package models

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
)

// Job for importing datasets
type Job struct {
	JobID         string          `json:"job_id,omitempty"`
	Recipe        string          `json:"recipe,omitempty"`
	State         string          `json:"state,omitempty"`
	UploadedFiles *[]UploadedFile `json:"files,omitempty"`
	Instances     []IDLink        `json:"instances,omitempty"`
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

// Validate the content of the structure
func (e Event) Validate() error {
	if e.MessageOffset == "" || e.Time == "" || e.Type == "" || e.Message == "" {
		return errors.New("Invalid event structure")
	}
	return nil
}

// Dimension which has been extracted from a dataset
type Dimension struct {
	Name   string `json:"dimension_id"`
	Value  string `json:"value"`
	NodeID string `json:"node_id"`
}

// Instance which presents a single dataset being imported
type Instance struct {
	InstanceID           string    `json:"instance_id,omitempty"`
	Job                  IDLink    `json:"job,omitempty"`
	State                string    `json:"state,omitempty"`
	Events               *[]Event  `json:"events,omitempty"`
	TotalObservations    *int      `json:"total_observations,omitempty"`
	InsertedObservations *int      `json:"total_inserted_observations,omitempty"`
	Headers              *[]string `json:"headers,omitempty"`
	LastUpdated          string    `json:"last_updated,omitempty"`
}

// UploadedFile used for a file which has been uploaded to a bucket
type UploadedFile struct {
	AliasName string `json:"alias_name" avro:"alias-name"`
	URL       string `json:"url" avro:"url"`
}

// Validate the content of the structure
func (s UploadedFile) Validate() error {
	if s.URL == "" || s.AliasName == "" {
		return errors.New("invalid json object received, alias_name and url are required")
	}
	return nil
}

// ImportData used to create a message to data baker or direct to the dimension-extractor
type ImportData struct {
	JobID         string
	Recipe        string         `json:"recipe,omitempty"`
	UploadedFiles []UploadedFile `json:"files,omitempty"`
	InstanceIDs   []string
}

// DataBakerEvent used to trigger the databaker process
type DataBakerEvent struct {
	JobID string `avro:"job_id"`
}

// UniqueDimensionValues hold all the unique values from a dimension
type UniqueDimensionValues struct {
	Name   string   `json:"dimension_id"`
	Values []string `json:"values"`
}

// IDLink holds the id and a link to the resource
type IDLink struct {
	ID   string `json:"id"`
	Link string `json:"link"`
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

// CreateEvent from a json message
func CreateEvent(reader io.Reader) (*Event, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}
	var message Event
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		return nil, errors.New("Failed to parse json body")
	}
	return &message, message.Validate()
}

// CreateInstance from a json message
func CreateInstance(reader io.Reader) (*Instance, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}
	var instance Instance
	err = json.Unmarshal(bytes, &instance)
	if err != nil {
		return nil, errors.New("Failed to parse json body")
	}
	return &instance, err
}
