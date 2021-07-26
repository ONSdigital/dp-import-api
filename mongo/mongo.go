package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	mongo "github.com/ONSdigital/dp-mongodb"
	mongolock "github.com/ONSdigital/dp-mongodb/dplock"
	mongohealth "github.com/ONSdigital/dp-mongodb/health"
	"github.com/ONSdigital/log.go/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var _ datastore.DataStorer = (*Mongo)(nil)

// Mongo represents a simplistic MongoDB configuration
type Mongo struct {
	Collection   string
	Database     string
	URI          string
	Session      *mgo.Session
	healthClient *mongohealth.CheckMongoClient
	lockClient   *mongolock.Lock
}

// NewDatastore creates a new mgo.Session with a strong consistency and a write mode of "majority"
func NewDatastore(ctx context.Context, uri, database, collection string) (m *Mongo, err error) {

	m = &Mongo{
		Collection: collection,
		Database:   database,
		URI:        uri}

	m.Session, err = mgo.Dial(uri)
	if err != nil {
		return nil, err
	}

	m.Session.EnsureSafe(&mgo.Safe{WMode: "majority"})
	m.Session.SetMode(mgo.Strong, true)

	importLocksCollection := fmt.Sprintf("%s_locks", collection)

	databaseCollectionBuilder := make(map[mongohealth.Database][]mongohealth.Collection)
	databaseCollectionBuilder[(mongohealth.Database)(database)] = []mongohealth.Collection{(mongohealth.Collection)(collection), (mongohealth.Collection)(importLocksCollection)}

	// Create client and healthclient from session
	client := mongohealth.NewClientWithCollections(m.Session, databaseCollectionBuilder)
	m.healthClient = &mongohealth.CheckMongoClient{
		Client:      *client,
		Healthcheck: client.Healthcheck,
	}

	m.lockClient = mongolock.New(ctx, m.Session, database, collection)

	return m, nil
}

// AcquireInstanceLock tries to lock the provided jobID.
// If the job is already locked, this function will block until it's released,
// at which point we acquire the lock and return.
// Note: the lock is currently only used to update processed_instances
func (m *Mongo) AcquireInstanceLock(ctx context.Context, jobID string) (lockID string, err error) {
	return m.lockClient.Acquire(ctx, jobID)
}

// UnlockInstance releases an exclusive mongoDB lock for the provided lockId (if it exists)
// Note: the lock is currently only used to update processed_instances
func (m *Mongo) UnlockInstance(lockID string) error {
	return m.lockClient.Unlock(lockID)
}

// GetJobs retrieves all import documents matching filters
func (m *Mongo) GetJobs(ctx context.Context, filters []string, offset int, limit int) (*models.JobResults, error) {
	s := m.Session.Copy()
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
	s := m.Session.Copy()
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
	s := m.Session.Copy()
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
	s := m.Session.Copy()
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
	s := m.Session.Copy()
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

// UpdateProcessedInstance overides the processed instances for an existing import job
func (m *Mongo) UpdateProcessedInstance(id string, procInstances []models.ProcessedInstances) (err error) {
	s := m.Session.Copy()
	defer s.Close()

	update := bson.M{
		"$set": bson.M{"processed_instances": procInstances},
		"$currentDate": bson.M{
			"last_updated": true,
			"unique_timestamp": bson.M{
				"$type": "timestamp",
			},
		},
	}

	err = s.DB(m.Database).C(m.Collection).Update(bson.M{"id": id}, update)

	if err != nil && err == mgo.ErrNotFound {
		return errs.ErrJobNotFound
	}

	return nil
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}

// Close disconnects the mongo session
func (m *Mongo) Close(ctx context.Context) error {
	return mongo.Close(ctx, m.Session)
}
