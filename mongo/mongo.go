package mongo

import (
	"context"
	"github.com/ONSdigital/log.go/log"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
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
func (m *Mongo) GetJobs(ctx context.Context, filters []string, offset int, limit int) (*models.JobResults, error) {
	s := session.Copy()
	defer s.Close()

	var stateFilter bson.M
	if len(filters) > 0 {
		stateFilter = bson.M{"state": bson.M{"$in": filters}}
	}
	query := s.DB(m.Database).C(m.Collection).Find(stateFilter)
	totalCount, err := query.Count()
	if err != nil {
		log.Event(ctx, "error counting items", log.ERROR, log.Error(err))
		if err == mgo.ErrNotFound {
			return &models.JobResults{
				Items:      []*models.Job{},
				Count:      0,
				TotalCount: 0,
				Offset:     offset,
				Limit:      limit,
			}, nil
		}
		return nil, err
	}
	if totalCount < 1 {
		return nil, errs.ErrJobNotFound
	}

	var jobItems []*models.Job
	if limit > 0 {
		iter := query.Sort().Skip(offset).Limit(limit).Iter()
		defer func() {
			err := iter.Close()
			if err != nil {
				log.Event(ctx, "error closing job iterator", log.ERROR, log.Error(err), log.Data{"filter": stateFilter})
			}
		}()

		if err := iter.All(&jobItems); err != nil {
			if err == mgo.ErrNotFound {
				return &models.JobResults{
					Items:      []*models.Job{},
					Count:      0,
					TotalCount: totalCount,
					Offset:     offset,
					Limit:      limit,
				}, nil
			}
			return nil, err
		}
	}

	return &models.JobResults{
		Items:      jobItems,
		Count:      len(jobItems),
		TotalCount: totalCount,
		Offset:     offset,
		Limit:      limit,
	}, nil
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

	return nil
}

// HealthCheckClient generates a healthcheck client for this mongoDB, with an existing session
func (m *Mongo) HealthCheckClient() *mongohealth.CheckMongoClient {

	databaseCollectionBuilder := make(map[mongohealth.Database][]mongohealth.Collection)
	databaseCollectionBuilder[(mongohealth.Database)(m.Database)] = []mongohealth.Collection{(mongohealth.Collection)(m.Collection)}

	client := mongohealth.NewClientWithCollections(session, databaseCollectionBuilder)

	return &mongohealth.CheckMongoClient{
		Client:      *client,
		Healthcheck: client.Healthcheck,
	}
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	checkMongoClient := m.HealthCheckClient()
	return checkMongoClient.Checker(ctx, state)
}

// Close disconnects the mongo session
func (m *Mongo) Close(ctx context.Context) error {
	return mongo.Close(ctx, session)
}
