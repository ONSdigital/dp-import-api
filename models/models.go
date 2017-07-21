package models

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// Job - A structure for an importqueue job
type Job struct {
	JobID         string          `json:"job_id,omitempty"`
	Recipe        string          `json:"recipe,omitempty"`
	State         string          `json:"state,omitempty"`
	UploadedFiles *[]UploadedFile `json:"files,omitempty"`
	Links         JobLinks        `json:"links,omitempty"`
}

type JobLinks struct {
	InstanceIDs []string `json:"instance_ids"`
}

// Validate - Validate the content of the structure
func (job *Job) Validate() error {
	if job.Recipe == "" {
		return fmt.Errorf("Missing properties to create importqueue job struct")
	}
	if job.State == "" {
		job.State = "created"
	}
	if job.UploadedFiles == nil {
		job.UploadedFiles = &[]UploadedFile{}
	}
	return nil
}

// Event - A structure for an event for an instance
type Event struct {
	Type          string `json:"type"`
	Time          string `json:"time"`
	Message       string `json:"message"`
	MessageOffset string `json:"messageOffset"`
}

// Validate - Validate the content of the structure
func (e Event) Validate() error {
	if e.MessageOffset == "" || e.Time == "" || e.Type == "" || e.Message == "" {
		return fmt.Errorf("Invalid event structure")
	}
	return nil
}

// Dimension - A structure for a dimension
type Dimension struct {
	Name   string `json:"dimension_name"`
	Value  string `json:"value"`
	NodeID string `json:"node_id"`
}

// JobInstanceState - A structure used for a instance job
type Instance struct {
	InstanceID           string   `json:"instance_id"`
	State                string   `json:"state"`
	Events               []Event  `json:"events"`
	NumberOfObservations int      `json:"number_of_observations"`
	Headers              []string `json:"headers"`
	LastUpdated          string   `json:"last_updated"`
}

// UploadedFile - a structure used for a file which has been uploaded to a bucket
type UploadedFile struct {
	AliasName string `json:"alias_name" avro:"alias-name"`
	URL       string `json:"url" avro:"url"`
}

// Validate - Validate the content of the structure
func (s UploadedFile) Validate() error {
	if s.URL == "" || s.AliasName == "" {
		return fmt.Errorf("Invalid s3 file structure")
	}
	return nil
}

// ImportData - A structure used to create a message to data baker
type ImportData struct {
	JobId         string
	Recipe        string         `json:"recipe,omitempty"`
	UploadedFiles []UploadedFile `json:"files,omitempty"`
	InstanceIds   []string
}

type DataBakerEvent struct {
	JobId string `avro:"job_id"`
}

// CreateJob - Create a job from a reader
func CreateJob(reader io.Reader) (*Job, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var job Job
	badJSONError := json.Unmarshal(bytes, &job)
	if badJSONError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &job, nil
}

// CreateUploadedFile - Create a uploaded file from a reader
func CreateUploadedFile(reader io.Reader) (*UploadedFile, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message UploadedFile
	badJSONError := json.Unmarshal(bytes, &message)
	if badJSONError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
}

// CreateEvent - Create an event from a reader
func CreateEvent(reader io.Reader) (*Event, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message Event
	badJSONError := json.Unmarshal(bytes, &message)
	if badJSONError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
}
