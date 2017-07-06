package models

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type Message interface {
	Validate() error
}

type ImportJob struct {
	Dataset string
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
	InstanceId string `json:"instanceId"`
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
	AliasName string
	S3Url     string
}

func (m ImportJob) Validate() error {
	if m.Dataset == "" {
		return fmt.Errorf("No dataset was provided")
	}
	return nil
}

func (s S3File) Validate() error {
	if s.S3Url == "" || s.AliasName == "" {
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
