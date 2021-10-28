package models

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	bsonprim "go.mongodb.org/mongo-driver/bson/primitive"
)

// CreatedState represents one possible state of the job resource
const (
	CompletedState = "completed"
	CreatedState   = "created"
	SubmittedState = "submitted"
	FailedState    = "failed"
)

var validStates = map[string]bool{
	CompletedState: true,
	CreatedState:   true,
	SubmittedState: true,
	FailedState:    true,
}

// JobResults for list of Job items
type JobResults struct {
	Count      int    `json:"count"`
	Offset     int    `json:"offset"`
	Limit      int    `json:"limit"`
	TotalCount int    `json:"total_count"`
	Items      []*Job `json:"items"`
}

// Job for importing datasets
type Job struct {
	ID              string               `bson:"id,omitempty"                  json:"id,omitempty"`
	RecipeID        string               `bson:"recipe,omitempty"              json:"recipe,omitempty"`
	State           string               `bson:"state,omitempty"               json:"state,omitempty"`
	UploadedFiles   *[]UploadedFile      `bson:"files,omitempty"               json:"files,omitempty"`
	Links           LinksMap             `bson:"links,omitempty"               json:"links,omitempty"`
	Processed       []ProcessedInstances `bson:"processed_instances,omitempty" json:"processed_instances,omitempty"`
	LastUpdated     time.Time            `bson:"last_updated,omitempty"        json:"last_updated,omitempty"`
	UniqueTimestamp bsonprim.Timestamp   `bson:"unique_timestamp,omitempty"    json:"-"`
}

// LinksMap represents a list of links related to a job resource
type LinksMap struct {
	Instances []IDLink `bson:"instances,omitempty" json:"instances,omitempty"`
	Self      IDLink   `bson:"self,omitempty" json:"self,omitempty"`
}

// Validate the content of a job
func (job *Job) Validate() error {
	if job.RecipeID == "" {
		return errs.ErrMissingProperties
	}
	if job.State == "" {
		job.State = CreatedState
	}
	if job.UploadedFiles == nil {
		job.UploadedFiles = &[]UploadedFile{}
	}
	return nil
}

// ValidateState checks the state is valid
func (job *Job) ValidateState() error {
	if job.State == "" {
		return nil
	}

	if !validStates[job.State] {
		return errs.ErrInvalidState
	}

	return nil
}

// UploadedFile used for a file which has been uploaded to a bucket
type UploadedFile struct {
	AliasName string `bson:"alias_name" json:"alias_name" avro:"alias-name"`
	URL       string `bson:"url"        json:"url"        avro:"url"`
}

// Validate the content of the structure
func (s UploadedFile) Validate() error {
	if s.URL == "" || s.AliasName == "" {
		return errs.ErrInvalidUploadedFileObject
	}
	return nil
}

// ImportData used to create a message to data baker or direct to the dimension-extractor
type ImportData struct {
	JobID         string
	Recipe        string          `json:"recipe,omitempty"`
	Format        string          `json:"format,omitempty"`
	UploadedFiles *[]UploadedFile `json:"files,omitempty"`
	InstanceIDs   []string
}

// DataBakerEvent used to trigger the databaker process
type DataBakerEvent struct {
	JobID string `avro:"job_id"`
}

// IDLink holds the ID and a link to the resource
type IDLink struct {
	ID   string `bson:"id,omitempty" json:"id"`
	HRef string `bson:"href,omitempty" json:"href"`
}

// ProcessedInstances holds the ID and the number of code lists that have been processed during an import process for an instance
type ProcessedInstances struct {
	ID             string `bson:"id,omitempty"               json:"id,omitempty"`
	RequiredCount  int    `bson:"required_count,omitempty"   json:"required_count,omitempty"`
	ProcessedCount int    `bson:"processed_count,omitempty"  json:"processed_count,omitempty"`
}

// CreateJob from a json message
func CreateJob(reader io.Reader) (*Job, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrFailedToReadRequestBody
	}
	var job Job
	err = json.Unmarshal(bytes, &job)
	if err != nil {
		return nil, errs.ErrFailedToParseJSONBody
	}
	return &job, nil
}

// CreateUploadedFile from a json message
func CreateUploadedFile(reader io.Reader) (*UploadedFile, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrFailedToReadRequestBody
	}
	var message UploadedFile
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		return nil, errs.ErrFailedToParseJSONBody
	}
	return &message, message.Validate()
}

// CreateInstance from a job ID and the provided recipe CodeLists
// Neither job nor job.Links can be nil
func CreateInstance(job *Job, datasetID, datasetURL string, codelists []recipe.CodeList) *dataset.NewInstance {

	buildHierarchyTasks := make([]*dataset.BuildHierarchyTask, 0)
	buildSearchTasks := make([]*dataset.BuildSearchIndexTask, 0)

	// Map from recipe codelists to dataset codelists and import tasks
	datasetCodeLists := make([]dataset.CodeList, len(codelists))
	for i, codelist := range codelists {

		isHierarchy := false
		if codelist.IsHierarchy != nil && *codelist.IsHierarchy {
			isHierarchy = true
		}

		datasetCodeLists[i] = dataset.CodeList{
			ID:          codelist.ID,
			HRef:        codelist.HRef,
			Name:        codelist.Name,
			IsHierarchy: isHierarchy,
		}

		if isHierarchy {
			buildHierarchyTasks = append(buildHierarchyTasks, &dataset.BuildHierarchyTask{
				State:         CreatedState,
				CodeListID:    codelist.ID,
				DimensionName: codelist.Name,
			})

			buildSearchTasks = append(buildSearchTasks, &dataset.BuildSearchIndexTask{
				State:         CreatedState,
				DimensionName: codelist.Name,
			})
		}
	}

	return &dataset.NewInstance{
		Dimensions: datasetCodeLists,
		Links: &dataset.Links{
			Job:     dataset.Link{ID: job.ID, URL: job.Links.Self.HRef},
			Dataset: dataset.Link{ID: datasetID, URL: datasetURL},
		},
		ImportTasks: &dataset.InstanceImportTasks{
			ImportObservations: &dataset.ImportObservationsTask{
				State:                CreatedState,
				InsertedObservations: 0,
			},
			BuildHierarchyTasks:   buildHierarchyTasks,
			BuildSearchIndexTasks: buildSearchTasks,
		},
	}
}
