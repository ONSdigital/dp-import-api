package mongo

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/models"
)

var InternalError = errors.New("DataStore internal error")

type DataStorer struct {
	NotFound      bool
	InternalError bool
}

func (ds *DataStorer) AddJob(importJob *models.Job) (*models.Job, error) {
	if ds.InternalError {
		return &models.Job{}, InternalError
	}
	return &models.Job{ID: "34534543543"}, nil
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
	return &models.Job{ID: "34534543543"}, nil
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

func (m *DataStorer) Close(ctx context.Context) error {
	return nil
}

func (m *DataStorer) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return nil
}
