package dataset

import (
	"context"

	"github.com/ONSdigital/dp-import-api/models"
)

type DatasetAPIer interface {
	CreateInstance(ctx context.Context, jobID, jobURL string) (*models.Instance, error)
	GetURL() string
	UpdateInstanceState(ctx context.Context, instanceID, newState string) error
}
