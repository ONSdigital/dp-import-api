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
	ID            string          `bson:"id,omitempty"             json:"id,omitempty"`
	RecipeID      string          `bson:"recipe,omitempty"         json:"recipe,omitempty"`
	State         string          `bson:"state,omitempty"          json:"state,omitempty"`
	UploadedFiles *[]UploadedFile `bson:"files,omitempty"          json:"files,omitempty"`
	Links         LinksMap        `bson:"links,omitempty"          json:"links,omitempty"`
	LastUpdated   time.Time       `bson:"last_updated,omitempty"   json:"last_updated,omitempty"`
}

type LinksMap struct {
	Instances []IDLink `bson:"instances,omitempty" json:"instances,omitempty"`
	Self      IDLink   `bson:"self,omitempty" json:"self,omitempty"`
}

// Validate the content of a job
func (job *Job) Validate() error {
	if job.RecipeID == "" {
		return errors.New("missing properties to create import queue job struct")
	}
	if job.State == "" {
		job.State = "created"
	}
	if job.UploadedFiles == nil {
		job.UploadedFiles = &[]UploadedFile{}
	}
	return nil
}

//this should probably be replaced with an import of
//github.com/ONSdigital/dp-code-list-api/{pkg when unstubbed}
type Recipe struct {
	ID              string           `json:"id"`
	Alias           string           `json:"alias"`
	OutputInstances []RecipeInstance `json:"output_instances"`
}

type RecipeInstance struct {
	DatasetID string     `json:"dataset_id"`
	CodeLists []CodeList `json:"code_lists"`
}

type CodeList struct {
	ID   string `json:"id"`
	HRef string `json:"href"`
	Name string `json:"name"`
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
	Dimensions           []CodeList     `json:"dimensions,omitempty"`
	LastUpdated          string         `json:"last_updated,omitempty"`
}

type InstanceLinks struct {
	Job     IDLink `json:"job,omitempty"`
	Dataset IDLink `json:"dataset,omitempty"`
}

// UploadedFile used for a file which has been uploaded to a bucket
type UploadedFile struct {
	AliasName string `bson:"alias_name" json:"alias_name" avro:"alias-name"`
	URL       string `bson:"url"        json:"url"        avro:"url"`
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
	Recipe        string          `json:"recipe,omitempty"`
	UploadedFiles *[]UploadedFile `json:"files,omitempty"`
	InstanceIDs   []string
}

// DataBakerEvent used to trigger the databaker process
type DataBakerEvent struct {
	JobID string `avro:"job_id"`
}

// IDLink holds the ID and a link to the resource
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
func CreateInstance(job *Job, datasetID, datasetURL string, codelists []CodeList) *Instance {
	return &Instance{
		Dimensions: codelists,
		Links: &InstanceLinks{
			Job:     IDLink{ID: job.ID, HRef: job.Links.Self.HRef},
			Dataset: IDLink{ID: datasetID, HRef: datasetURL},
		}}
}
