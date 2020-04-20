package mongo

import (
	"context"
	"time"

	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	mongo "github.com/ONSdigital/dp-mongodb"
	mongohealth "github.com/ONSdigital/dp-mongodb/health"
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
			return nil, errs.ErrJobNotFound
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
			return nil, errs.ErrJobNotFound
		}
		return nil, err
	}
	return &job, nil
}

// AddJob adds an ImportJob document
func (m *Mongo) AddJob(job *models.Job) (*models.Job, error) {
	s := session.Copy()
	defer s.Close()

	currentTime := time.Now().UTC()
	job.LastUpdated = currentTime
	// Replace line below with
	// job.UniqueTimestamp = bson.NewMongoTimestamp(currentTime, 1)
	// once mgo has been updated with new function `NewMongoTimestamp`
	job.UniqueTimestamp = 1

	if err := s.DB(m.Database).C(m.Collection).Insert(job); err != nil {
		return nil, err
	}

	return m.GetJob(job.ID)
}

// AddUploadedFile adds an UploadedFile to an import job
func (m *Mongo) AddUploadedFile(id string, file *models.UploadedFile) error {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$addToSet": bson.M{
			"files": bson.M{
				"alias_name": file.AliasName,
				"url":        file.URL,
			},
		},
		"$currentDate": bson.M{
			"last_updated": true,
			"unique_timestamp": bson.M{
				"$type": "timestamp",
			},
		},
	}

	// Replace above with below once go-ns mongo package has been updated
	// update := bson.M{
	// 	"$addToSet": bson.M{
	// 		"files": bson.M{
	// 			"alias_name": file.AliasName,
	// 			"url":        file.URL,
	// 		},
	// 	},
	// }
	// mongo.WithUpdates(update)
	err := s.DB(m.Database).C(m.Collection).Update(bson.M{"id": id}, update)
	if err != nil && err == mgo.ErrNotFound {
		return errs.ErrJobNotFound
	}

	return nil
}

// UpdateJob adds or overides an existing import job
func (m *Mongo) UpdateJob(id string, job *models.Job) (err error) {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$set": job,
		"$currentDate": bson.M{
			"last_updated": true,
			"unique_timestamp": bson.M{
				"$type": "timestamp",
			},
		},
	}

	// Replace above with below once go-ns mongo package has been updated
	//mongo.WithUpdates(bson.M{"$set": job})
	err = s.DB(m.Database).C(m.Collection).Update(bson.M{"id": id}, update)

	if err != nil && err == mgo.ErrNotFound {
		return errs.ErrJobNotFound
	}

	return
}

// UpdateJobState changes the state attribute of an import job
func (m *Mongo) UpdateJobState(id, newState string) (err error) {
	s := session.Copy()
	defer s.Close()

	update := bson.M{
		"$set": bson.M{"state": newState},
		"$currentDate": bson.M{
			"last_updated": true,
			"unique_timestamp": bson.M{
				"$type": "timestamp",
			},
		},
	}

	// Replace above with below once go-ns mongo package has been updated
	// mongo.WithUpdates(bson.M{"$set": bson.M{"state": newState}})
	_, err = s.DB(m.Database).C(m.Collection).Upsert(bson.M{"id": id}, update)
	return
}

func (m *Mongo) HealthCheckClient() *mongohealth.CheckMongoClient {
	client := mongohealth.NewClient(session)

	return &mongohealth.CheckMongoClient{
		Client:      *client,
		Healthcheck: client.Healthcheck,
	}
}

// Close disconnects the mongo session
func (m *Mongo) Close(ctx context.Context) error {
	return mongo.Close(ctx, session)
}
