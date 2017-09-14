package dataset

import "github.com/ONSdigital/dp-import-api/models"

type DatasetAPI struct {
	url string
}

func CreateDatasetAPI() *DatasetAPI {
	return &DatasetAPI{url: "http://..."}
}

func (ds *DatasetAPI) GetURL() string {
	return ds.url
}

func (ds *DatasetAPI) CreateInstance(string, string) (*models.Instance, error) {
	return &models.Instance{}, nil
}

func (ds *DatasetAPI) UpdateInstanceState(string, string) error {
	return nil
}
