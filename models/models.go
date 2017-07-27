package models

import (
	"encoding/json"
	"io"
	"errors"
	"io/ioutil"
	"fmt"
)

// A structure for an importqueue job
type Job struct {
	JobID         string          `json:"job_id,omitempty"`
	Recipe        string          `json:"recipe,omitempty"`
	State         string          `json:"state,omitempty"`
	UploadedFiles *[]UploadedFile `json:"files,omitempty"`
	Links         JobLinks        `json:"links,omitempty"`
}

// A list of links to instance jobs
type JobLinks struct {
	InstanceIDs []string `json:"instance_ids"`
}

//Validate the content of the structure
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

// A structure for an event for an instance
type Event struct {
	Type          string `json:"type"`
	Time          string `json:"time"`
	Message       string `json:"message"`
	MessageOffset string `json:"messageOffset"`
}

// Validate the content of the structure
func (e Event) Validate() error {
	if e.MessageOffset == "" || e.Time == "" || e.Type == "" || e.Message == "" {
		return fmt.Errorf("Invalid event structure")
	}
	return nil
}

// A structure for a dimension
type Dimension struct {
	Name   string `json:"dimension_name"`
	Value  string `json:"value"`
	NodeID string `json:"node_id"`
}

// A structure used for a instance job
type Instance struct {
	InstanceID           string    `json:"instance_id,omitempty"`
	State                string    `json:"state,omitempty"`
	Events               *[]Event  `json:"events,omitempty"`
	NumberOfObservations int       `json:"number_of_observations,omitempty"`
	Headers              *[]string `json:"headers,omitempty"`
	LastUpdated          string    `json:"last_updated,omitempty"`
}

//  A structure used for a file which has been uploaded to a bucket
type UploadedFile struct {
	AliasName string `json:"alias_name" avro:"alias-name"`
	URL       string `json:"url" avro:"url"`
}

// Validate the content of the structure
func (s UploadedFile) Validate() error {
	if s.URL == "" || s.AliasName == "" {
		return fmt.Errorf("Invalid s3 file structure")
	}
	return nil
}

// A structure used to create a message to data baker
type ImportData struct {
	JobID         string
	Recipe        string         `json:"recipe,omitempty"`
	UploadedFiles []UploadedFile `json:"files,omitempty"`
	InstanceIDs   []string
}

// An event used to trigger the databaker process
type DataBakerEvent struct {
	JobID string `avro:"job_id"`
}

// Create a job from a reader
func CreateJob(reader io.Reader) (*Job, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var job Job
	err = json.Unmarshal(bytes, &job)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &job, nil
}

// Create a uploaded file from a reader
func CreateUploadedFile(reader io.Reader) (*UploadedFile, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message UploadedFile
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
}

// Create an event from a reader
func CreateEvent(reader io.Reader) (*Event, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message Event
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
}

// Create an instance
func CreateInstance(reader io.Reader) (*Instance, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var instance Instance
	err = json.Unmarshal(bytes, &instance)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &instance, err
}
