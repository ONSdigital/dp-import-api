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
var CreatedJob = models.Job{ID: "34534543543"}

func (ds *DataStorer) AddJob(_ context.Context, _ *models.Job) (*models.Job, error) {
	if ds.InternalError {
		return &models.Job{}, InternalError
	}
	return &CreatedJob, nil
}

func (ds *DataStorer) GetJobs(_ context.Context, _ []string, _ int, _ int) (*models.JobResults, error) {
	if ds.InternalError {
		return &models.JobResults{Items: []*models.Job{}}, InternalError
	}
	return &models.JobResults{Items: []*models.Job{{ID: "34534543543"}}}, nil
}

func (ds *DataStorer) GetJob(_ context.Context, _ string) (*models.Job, error) {
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

func (ds *DataStorer) AddInstance(_ context.Context, _ string) (string, error) {
	if ds.NotFound {
		return "", errs.ErrJobNotFound
	}
	if ds.InternalError {
		return "", InternalError
	}
	return "123", nil
}

func (ds *DataStorer) UpdateJob(_ context.Context, _ string, _ *models.Job) error {
	if ds.NotFound {
		return errs.ErrJobNotFound
	}
	if ds.InternalError {
		return InternalError
	}
	return nil
}

func (ds *DataStorer) UpdateJobState(_ context.Context, _ string, _ string) error {
	if ds.NotFound {
		return errs.ErrJobNotFound
	}
	if ds.InternalError {
		return InternalError
	}
	return nil
}

func (ds *DataStorer) AddUploadedFile(_ context.Context, _ string, _ *models.UploadedFile) error {
	if ds.NotFound {
		return errs.ErrJobNotFound
	}
	if ds.InternalError {
		return InternalError
	}
	return nil
}

func (ds *DataStorer) UpdateProcessedInstance(_ context.Context, _ string, _ []models.ProcessedInstances) (err error) {
	if ds.NotFound {
		return errs.ErrJobNotFound
	}
	if ds.InternalError {
		return InternalError
	}
	return nil
}

func (ds *DataStorer) Close(_ context.Context) error {
	return nil
}

func (ds *DataStorer) Checker(_ context.Context, _ *healthcheck.CheckState) error {
	return nil
}

func (ds *DataStorer) AcquireInstanceLock(_ context.Context, _ string) (lockID string, err error) {
	if ds.IsLocked {
		return "", errors.New("already locked")
	}
	ds.IsLocked = true
	ds.HasBeenLocked = true
	return testLockID, nil
}

func (ds *DataStorer) UnlockInstance(_ context.Context, lockID string) {
	if lockID == testLockID {
		ds.IsLocked = false
	}
}
