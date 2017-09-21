package dataset

import (
	"context"

	"github.com/ONSdigital/dp-import-api/models"
)

type DatasetAPI struct {
	url string
}

func CreateDatasetAPI() *DatasetAPI {
	return &DatasetAPI{url: "http://..."}
}

func (ds *DatasetAPI) GetURL() string {
	return ds.url
}

func (ds *DatasetAPI) CreateInstance(context.Context, string, string) (*models.Instance, error) {
	return &models.Instance{}, nil
}

func (ds *DatasetAPI) UpdateInstanceState(context.Context, string, string) error {
	return nil
}
