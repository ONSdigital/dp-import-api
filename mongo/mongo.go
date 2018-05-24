package mongo

import (
	"context"
	"time"

	"github.com/ONSdigital/dp-import-api/api-errors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	mongolib "github.com/ONSdigital/go-ns/mongo"

	"github.com/gedge/mgo"
	"github.com/gedge/mgo/bson"
)

var _ datastore.DataStorer = (*Mongo)(nil)

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
	defer s.Close()
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

// AddJob adds an ImportJob document
func (m *Mongo) AddJob(job *models.Job) (*models.Job, error) {
	s := session.Copy()
	defer s.Close()
	job.LastUpdated = time.Now().UTC()
	if err := s.DB(m.Database).C(m.Collection).Insert(job); err != nil {
		return nil, err
	}

	return m.GetJob(job.ID)
}

// AddUploadedFile adds an UploadedFile to an import job
func (m *Mongo) AddUploadedFile(id string, file *models.UploadedFile) error {
	s := session.Copy()
	defer s.Close()

	update := mongolib.WithUpdates(bson.M{
		"$addToSet": bson.M{
			"files": bson.M{
				"alias_name": file.AliasName,
				"url":        file.URL,
			},
		},
	})

	err := s.DB(m.Database).C(m.Collection).Update(bson.M{"id": id}, update)
	if err != nil && err == mgo.ErrNotFound {
		return api_errors.JobNotFoundError
	}
	return nil

}

// UpdateJob adds or overides an existing import job
func (m *Mongo) UpdateJob(id string, job *models.Job) (err error) {
	s := session.Copy()
	defer s.Close()

	update := mongolib.WithUpdates(bson.M{
		"$set": job,
	})

	err = s.DB(m.Database).C(m.Collection).Update(bson.M{"id": id}, update)

	if err != nil && err == mgo.ErrNotFound {
		return api_errors.JobNotFoundError
	}

	return
}

// UpdateJobState changes the state attribute of an import job
func (m *Mongo) UpdateJobState(id, newState string) (err error) {
	s := session.Copy()
	defer s.Close()

	update := mongolib.WithUpdates(bson.M{
		"$set": bson.M{"state": newState},
	})

	_, err = s.DB(m.Database).C(m.Collection).Upsert(bson.M{"id": id}, update)
	return
}

func (m *Mongo) Close(ctx context.Context) error {
	return mongolib.Close(ctx, session)
}
