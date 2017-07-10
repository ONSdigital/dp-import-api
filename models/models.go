package models

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)


// Job - A structure for an import job
type Job struct {
	Recipe        string         `json:"recipe"`
	State         string         `json:"state"`
	Datasets      []string       `json:"datasets"`
	UploadedFileS []UploadedFile `json:"s3Files"`
}

// NewJob - The requested structure to create a new job
type NewJob struct {
	Recipe   string   `json:"recipe"`
	Datasets []string `json:"datasets"`
}

// Validate - Validate the content of the structure
func (i NewJob) Validate() error {
	if i.Recipe == "" || i.Datasets == nil {
		return fmt.Errorf("Missing properties to create import job struct")
	}
	return nil
}

// JobState - The requested structure to update a jobs state
type JobState struct {
	State string `json:"state"`
}

// Validate - Validate the content of the structure
func (j JobState) Validate() error {
	if j.State == "" {
		return fmt.Errorf("No state provided")
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
	NodeName string `json:"nodeName"`
	Value    string `json:"value"`
	NodeID   string `json:"nodeId"`
}

// Validate - Validate the content of the structure
func (d Dimension) Validate() error {
	return nil
}

// JobInstance - A structure used to return to a client when a new job has been created
type JobInstance struct {
	JobID       string   `json:"jobId"`
	InstanceIds []string `json:"instanceIds"`
}

// JobInstanceState - A structure used for a instance job
type JobInstanceState struct {
	InstanceID  string  `json:"instanceId"`
	Dataset     string  `json:"dataset"`
	State       string  `json:"state"`
	Events      []Event `json:"events"`
	LastUpdated string  `json:"lastUpdated"`
}

// UploadedFile - a structure used for a file which has been uploaded to a bucket
type UploadedFile struct {
	AliasName string `json:"aliasName"`
	URL       string `json:"url"`
}

// Validate - Validate the content of the structure
func (s UploadedFile) Validate() error {
	if s.URL == "" || s.AliasName == "" {
		return fmt.Errorf("Invalid s3 file structure")
	}
	return nil
}

// PublishDataset - A structure used to create a message to data baker
type PublishDataset struct {
	Recipe        string         `json:"recipe"`
	UploadedFiles []UploadedFile `json:"files"`
	InstanceIds   []string       `json:"instanceIds"`
}

// CreateJobState - Create a job state from a reader
func CreateJobState(reader io.Reader) (*JobState, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var state JobState
	badJSONError := json.Unmarshal(bytes, &state)
	if badJSONError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &state, state.Validate()
}

// CreateJob - Create a job from a reader
func CreateJob(reader io.Reader) (*NewJob, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message NewJob
	badJSONError := json.Unmarshal(bytes, &message)
	if badJSONError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
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

// CreateDimension - Create a dimension from a reader
func CreateDimension(reader io.Reader) (*Dimension, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message Dimension
	badJSONError := json.Unmarshal(bytes, &message)
	if badJSONError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
}
