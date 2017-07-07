package models

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type Job struct {
	Recipe   string   `json:"recipe"`
	State    string   `json:"state"`
	Datasets []string `json:"datasets"`
	S3Files  []S3File `json:"s3Files"`
}

type ImportJob struct {
	Recipe   string   `json:"recipe"`
	Datasets []string `json:"datasets"`
}

type JobState struct {
	State string `json:"recipe"`
}

type Event struct {
	Type          string `json:"type"`
	Time          string `json:"time"`
	Message       string `json:"message"`
	MessageOffset string `json:"messageOffset"`
}

type Dimension struct {
	NodeName string `json:"nodeName"`
	Value    string `json:"value"`
	NodeId   string `json:"nodeId"`
}

type JobInstance struct {
	JobId       string   `json:"jobId"`
	InstanceIds []string `json:"instanceIds"`
}

type ImportJobState struct {
	InstanceId  string   `json:"instanceId"`
	Dataset     string   `json:"dataset"`
	S3Files     []S3File `json:"s3Files"`
	State       string   `json:"state"`
	Events      []Event  `json:"events"`
	LastUpdated string   `json:"lastUpdated"`
}

type S3File struct {
	AliasName string `json:"aliasName"`
	Url       string `json:"url"`
}

func (i ImportJob) Validate() error {
	if i.Recipe == "" || i.Datasets == nil {
		return fmt.Errorf("Missing properties to create import job struct")
	}
	return nil
}

func (s S3File) Validate() error {
	if s.Url == "" || s.AliasName == "" {
		return fmt.Errorf("Invalid s3 file structure")
	}
	return nil
}

func (d Dimension) Validate() error {
	return nil
}

func (e Event) Validate() error {
	if e.MessageOffset == "" || e.Time == "" || e.Type == "" || e.Message == "" {
		return fmt.Errorf("Invalid event structure")
	}
	return nil
}

func CreateImportJob(reader io.Reader) (*ImportJob, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message ImportJob
	badJsonError := json.Unmarshal(bytes, &message)
	if badJsonError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
}

func CreateS3File(reader io.Reader) (*S3File, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message S3File
	badJsonError := json.Unmarshal(bytes, &message)
	if badJsonError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
}

func CreateEvent(reader io.Reader) (*Event, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message Event
	badJsonError := json.Unmarshal(bytes, &message)
	if badJsonError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
}

func CreateDimension(reader io.Reader) (*Dimension, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read message body")
	}
	var message Dimension
	badJsonError := json.Unmarshal(bytes, &message)
	if badJsonError != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}
	return &message, message.Validate()
}
