package mongo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-import-api/dataset/interface"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rhttp"
	uuid "github.com/satori/go.uuid"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var session *mgo.Session

// Mongo represents a simplistic MongoDB configuration
type Mongo struct {
	Collection string
	Database   string
	URI        string
}

// NewDatastore creates a new mgo.Session with a strong consistency and a write mode of "majority"
func NewDatastore(uri, database, collection string) (*Mongo, error) {
	if session == nil {
		var err error
		if session, err = mgo.Dial(uri); err != nil {
			return nil, err
		}

		session.EnsureSafe(&mgo.Safe{WMode: "majority"})
		session.SetMode(mgo.Strong, true)
	}
	return &Mongo{Collection: collection, Database: database, URI: uri}, nil
}

// GetJobs retrieves all import documents matching filters
func (m *Mongo) GetJobs(filters []string) ([]models.Job, error) {
	s := session.Copy()
	defer s.Close()

	var stateFilter bson.M
	if len(filters) > 0 {
		stateFilter = bson.M{"state": bson.M{"$in": filters}}
	}
	iter := s.DB(m.Database).C(m.Collection).Find(stateFilter).Iter()

	results := []models.Job{}
	if err := iter.All(&results); err != nil {
		if err == mgo.ErrNotFound {
			return nil, api_errors.JobNotFoundError
		}
		return nil, err
	}

	return results, nil
}

// GetJob retrieves a single import job
func (m *Mongo) GetJob(id string) (*models.Job, error) {
	s := session.Copy()
	defer s.Clone()
	var job models.Job
	err := s.DB(m.Database).C(m.Collection).Find(bson.M{"id": id}).One(&job)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, api_errors.JobNotFoundError
		}
		return nil, err
	}
	return &job, nil
}

// callDatasetAPI contacts the Dataset API returns the json body (action = PUT, GET, POST, ...)
func getRecipe(path string) (*models.Recipe, error) {

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.ErrorC("Failed to create request for RecipeAPI", err, nil)
		return nil, err
	}

	var cli *rhttp.Client
	resp, err := cli.Do(req)
	if err != nil {
		log.ErrorC("Failed to action RecipeAPI", err, nil)
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Debug("unexpected status code from API", nil)
	}

	defer resp.Body.Close()
	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC("Failed to read body from DatasetAPI", err, nil)
		return nil, err
	}

	var recipe *models.Recipe
	json.Unmarshal(jsonBody, &recipe)
	return recipe, nil
}

// AddJob adds an ImportJob document
func (m *Mongo) AddJob(importJob *models.Job, selfURL string, datasetAPI dataset.DatasetAPIer) (*models.Job, error) {
	s := session.Copy()
	defer s.Close()

	// Create unique id
	importJob.JobID = (uuid.NewV4()).String()
	selfURL = strings.Replace(selfURL, "{id}", importJob.JobID, -1)

	//Get details needed for instances from Recipe API
	recipe, err := getRecipe(importJob.RecipeURL)
	if err != nil {
		log.ErrorC("failed to get recipe details", err, nil)
		return nil, err
	}

	for _, oi := range recipe.OutputInstances {
		// now create an instance for this file
		instance, err := datasetAPI.CreateInstance(importJob.JobID, selfURL, &oi)
		if err != nil {
			return nil, err
		}
		importJob.Links.Instances = append(importJob.Links.Instances,
			models.IDLink{
				ID:   instance.InstanceID,
				HRef: datasetAPI.GetURL() + "/instances/" + instance.InstanceID,
			},
		)
	}

	if err = s.DB(m.Database).C(m.Collection).Insert(importJob); err != nil {
		return nil, err
	}

	return m.GetJob(importJob.JobID)
}

// AddUploadedFile adds an UploadedFile to an import job
func (m *Mongo) AddUploadedFile(id string, file *models.UploadedFile, datasetAPI dataset.DatasetAPIer, selfURL string) (instance *models.Instance, err error) {
	s := session.Copy()
	defer s.Close()

	// create an instance for this import job
	instance, err = datasetAPI.CreateInstance(id, selfURL)
	if err != nil {
		return nil, err
	}

	update := bson.M{
		"$addToSet": bson.M{
			"files": bson.M{
				"alias_name": file.AliasName,
				"url":        file.URL,
			},
			"links.instances": bson.M{
				"id":   instance.InstanceID,
				"href": datasetAPI.GetURL() + "/instances/" + instance.InstanceID,
			},
		},
		"$currentDate": bson.M{"last_updated": true},
	}

	if _, err = s.DB(m.Database).C(m.Collection).Upsert(bson.M{"id": id}, update); err != nil {
		return
	}
	return
}

// UpdateJob adds or overides an existing import job
func (m *Mongo) UpdateJob(id string, job *models.Job, withoutRestrictions bool) (err error) {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$set":         job,
		"$currentDate": bson.M{"last_updated": true},
	}

	_, err = s.DB(m.Database).C(m.Collection).Upsert(bson.M{"id": id}, update)
	return
}

// UpdateJobState changes the state attribute of an import job
func (m *Mongo) UpdateJobState(id, newState string, withoutRestrictions bool) (err error) {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$set":         bson.M{"state": newState},
		"$currentDate": bson.M{"last_updated": true},
	}

	_, err = s.DB(m.Database).C(m.Collection).Upsert(bson.M{"id": id}, update)
	return
}

// PrepareJob returns a format ready to send to downstream services via kafka
func (m *Mongo) PrepareJob(datasetAPI dataset.DatasetAPIer, jobID string) (*models.ImportData, error) {
	s := session.Copy()
	defer s.Close()

	importJob, err := m.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	instanceIds := make([]string, 0)
	for _, instanceRef := range importJob.Links.Instances {
		instanceIds = append(instanceIds, instanceRef.ID)

		if err = datasetAPI.UpdateInstanceState(instanceRef.ID, "submitted"); err != nil {
			return nil, err
		}
	}

	return &models.ImportData{
		JobID:         jobID,
		Recipe:        importJob.RecipeURL,
		UploadedFiles: importJob.UploadedFiles,
		InstanceIDs:   instanceIds,
	}, nil
}
