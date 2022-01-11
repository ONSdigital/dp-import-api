package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/log.go/v2/log"

	mongolock "github.com/ONSdigital/dp-mongodb/v3/dplock"
	mongohealth "github.com/ONSdigital/dp-mongodb/v3/health"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	bsonprim "go.mongodb.org/mongo-driver/bson/primitive"
)

var _ datastore.DataStorer = (*Mongo)(nil)

type Mongo struct {
	mongodriver.MongoDriverConfig

	connection   *mongodriver.MongoConnection
	healthClient *mongohealth.CheckMongoClient
	lockClient   *mongolock.Lock
}

// NewDatastore creates a new mongodb.MongoConnection with the given configuration
func NewDatastore(ctx context.Context, cfg config.MongoConfig) (m *Mongo, err error) {
	m = &Mongo{MongoDriverConfig: cfg}

	m.connection, err = mongodriver.Open(&m.MongoDriverConfig)
	if err != nil {
		return nil, err
	}

	databaseCollectionBuilder := map[mongohealth.Database][]mongohealth.Collection{
		mongohealth.Database(m.Database): {
			mongohealth.Collection(m.ActualCollectionName(config.ImportsCollection)),
			mongohealth.Collection(m.ActualCollectionName(config.ImportsLockCollection)),
		},
	}
	m.healthClient = mongohealth.NewClientWithCollections(m.connection, databaseCollectionBuilder)
	m.lockClient = mongolock.New(ctx, m.connection, config.ImportsCollection)

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
func (m *Mongo) UnlockInstance(ctx context.Context, lockID string) {
	m.lockClient.Unlock(ctx, lockID)
}

// GetJobs retrieves all import documents matching filters
func (m *Mongo) GetJobs(ctx context.Context, filters []string, offset int, limit int) (*models.JobResults, error) {
	stateFilter := bson.M{}
	if len(filters) > 0 {
		stateFilter["state"] = bson.M{"$in": filters}
	}

	var jobItems []*models.Job
	totalCount, err := m.connection.Collection(m.ActualCollectionName(config.ImportsCollection)).Find(ctx, stateFilter, &jobItems,
		mongodriver.Sort(bson.M{"_id": 1}), mongodriver.Offset(offset), mongodriver.Limit(limit))
	if err != nil {
		log.Error(ctx, "error finding items", err)
		return nil, err
	}
	if totalCount < 1 {
		return nil, apierrors.ErrJobNotFound
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
func (m *Mongo) GetJob(ctx context.Context, id string) (*models.Job, error) {
	var job models.Job
	if err := m.connection.Collection(m.ActualCollectionName(config.ImportsCollection)).FindOne(ctx, bson.M{"id": id}, &job); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return nil, apierrors.ErrJobNotFound
		}
		return nil, err
	}

	return &job, nil
}

// AddJob adds an ImportJob document - the ID is assumed to be set
func (m *Mongo) AddJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	currentTime := time.Now().UTC()
	job.LastUpdated = currentTime
	job.UniqueTimestamp = bsonprim.Timestamp{T: uint32(time.Now().Unix())}

	if _, err := m.connection.Collection(m.ActualCollectionName(config.ImportsCollection)).Insert(ctx, job); err != nil {
		return nil, err
	}

	return m.GetJob(ctx, job.ID)
}

// updateByID is a helper function to update a job given an update operator
func (m *Mongo) updateByID(ctx context.Context, id string, update bson.M) (err error) {
	if _, err = m.connection.Collection(m.ActualCollectionName(config.ImportsCollection)).Must().Update(ctx, bson.M{"id": id}, update); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return apierrors.ErrJobNotFound
		}
		return err
	}

	return nil
}

// AddUploadedFile adds an UploadedFile to an import job
func (m *Mongo) AddUploadedFile(ctx context.Context, id string, file *models.UploadedFile) error {
	return m.updateByID(ctx, id, bson.M{
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
	})
}

// UpdateJob adds or overides an existing import job
func (m *Mongo) UpdateJob(ctx context.Context, id string, job *models.Job) (err error) {
	return m.updateByID(ctx, id, bson.M{
		"$set": job,
		"$currentDate": bson.M{
			"last_updated": true,
			"unique_timestamp": bson.M{
				"$type": "timestamp",
			},
		},
	})
}

// UpdateProcessedInstance overides the processed instances for an existing import job
func (m *Mongo) UpdateProcessedInstance(ctx context.Context, id string, procInstances []models.ProcessedInstances) (err error) {
	return m.updateByID(ctx, id, bson.M{
		"$set": bson.M{"processed_instances": procInstances},
		"$currentDate": bson.M{
			"last_updated": true,
			"unique_timestamp": bson.M{
				"$type": "timestamp",
			},
		},
	})
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}

// Close disconnects the mongo session
func (m *Mongo) Close(ctx context.Context) error {
	return m.connection.Close(ctx)
}
