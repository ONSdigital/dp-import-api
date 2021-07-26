package mongo

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/models"
)

const testLockID = "testLockID"

var InternalError = errors.New("DataStore internal error")

type DataStorer struct {
	NotFound      bool
	InternalError bool
	IsLocked      bool
	HasBeenLocked bool
}

// CreatedJob represents a job returned by AddJob
var CreatedJob models.Job = models.Job{ID: "34534543543"}

func (ds *DataStorer) AddJob(importJob *models.Job) (*models.Job, error) {
	if ds.InternalError {
		return &models.Job{}, InternalError
	}
	return &CreatedJob, nil
}

func (ds *DataStorer) GetJobs(ctx context.Context, filter []string, offset int, limit int) (*models.JobResults, error) {
	if ds.InternalError {
		return &models.JobResults{Items: []*models.Job{}}, InternalError
	}
	return &models.JobResults{Items: []*models.Job{{ID: "34534543543"}}}, nil
}

func (ds *DataStorer) GetJob(jobID string) (*models.Job, error) {
	if ds.InternalError {
		return &models.Job{}, InternalError
	}
	if ds.NotFound {
		return &models.Job{}, errs.ErrJobNotFound
	}
	return &models.Job{
		ID: "34534543543",
		Processed: []models.ProcessedInstances{
			{
				ID:             "54321",
				RequiredCount:  5,
				ProcessedCount: 0,
			},
		},
	}, nil
}

func (ds *DataStorer) AddInstance(jobID string) (string, error) {
	if ds.NotFound {
		return "", errs.ErrJobNotFound
	}
	if ds.InternalError {
		return "", InternalError
	}
	return "123", nil
}

func (ds *DataStorer) UpdateJob(string, *models.Job) error {
	if ds.NotFound {
		return errs.ErrJobNotFound
	}
	if ds.InternalError {
		return InternalError
	}
	return nil
}

func (ds *DataStorer) UpdateJobState(string, string) error {
	if ds.NotFound {
		return errs.ErrJobNotFound
	}
	if ds.InternalError {
		return InternalError
	}
	return nil
}

func (ds *DataStorer) AddUploadedFile(jobID string, message *models.UploadedFile) error {
	if ds.NotFound {
		return errs.ErrJobNotFound
	}
	if ds.InternalError {
		return InternalError
	}
	return nil
}

func (ds *DataStorer) UpdateProcessedInstance(id string, procInstances []models.ProcessedInstances) (err error) {
	if ds.NotFound {
		return errs.ErrJobNotFound
	}
	if ds.InternalError {
		return InternalError
	}
	return nil
}

func (m *DataStorer) Close(ctx context.Context) error {
	return nil
}

func (m *DataStorer) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return nil
}

func (m *DataStorer) AcquireInstanceLock(ctx context.Context, jobID string) (lockID string, err error) {
	if m.IsLocked {
		return "", errors.New("already locked")
	}
	m.IsLocked = true
	m.HasBeenLocked = true
	return testLockID, nil
}

func (m *DataStorer) UnlockInstance(lockID string) error {
	if lockID == testLockID {
		m.IsLocked = false
		return nil
	}
	return errors.New("wrongLockID")
}
