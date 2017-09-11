package mocks

import (
	"errors"

	"github.com/ONSdigital/dp-filter-api/models"
)

var (
	internalServerError  = errors.New("DataStore internal error")
	unauthorisedError    = errors.New("Unauthorised")
	badRequestError      = errors.New("Bad request")
	forbiddenError       = errors.New("Forbidden")
	notFoundError        = errors.New("Not found")
	dimensionionNotFound = errors.New("Dimension not found")
	optionNotFound       = errors.New("Option not found")
)

type DataStore struct {
	NotFound          bool
	DimensionNotFound bool
	OptionNotFound    bool
	BadRequest        bool
	Forbidden         bool
	Unauthorised      bool
	InternalError     bool
}

func (ds *DataStore) AddFilter(host string, filterJob *models.Filter) (models.Filter, error) {
	if ds.InternalError {
		return models.Filter{}, internalServerError
	}
	return models.Filter{DatasetFilterID: "12345678"}, nil
}

func (ds *DataStore) AddFilterDimension(dimension *models.AddDimension) error {
	if ds.InternalError {
		return internalServerError
	}

	if ds.NotFound {
		return notFoundError
	}

	if ds.Forbidden {
		return forbiddenError
	}

	return nil
}

func (ds *DataStore) AddFilterDimensionOption(dimension *models.AddDimensionOption) error {
	if ds.InternalError {
		return internalServerError
	}

	if ds.NotFound {
		return notFoundError
	}

	if ds.BadRequest {
		return badRequestError
	}

	if ds.Forbidden {
		return forbiddenError
	}

	return nil
}

func (ds *DataStore) GetFilter(filterID string) (models.Filter, error) {
	if ds.NotFound {
		return models.Filter{}, notFoundError
	}

	if ds.InternalError {
		return models.Filter{}, internalServerError
	}
	return models.Filter{DatasetFilterID: "12345678"}, nil
}

func (ds *DataStore) GetFilterDimensions(filterID string) ([]models.Dimension, error) {
	dimensions := []models.Dimension{}

	if ds.NotFound {
		return nil, notFoundError
	}

	if ds.InternalError {
		return nil, internalServerError
	}

	dimensions = append(dimensions, models.Dimension{Name: "1_age", DimensionURL: "/filters/123/dimensions/1_age"})

	return dimensions, nil
}

func (ds *DataStore) GetFilterDimension(filterID string, name string) error {
	if ds.DimensionNotFound {
		return dimensionionNotFound
	}

	if ds.BadRequest {
		return badRequestError
	}

	if ds.InternalError {
		return internalServerError
	}

	return nil
}

func (ds *DataStore) GetFilterDimensionOptions(filterID string, name string) ([]models.DimensionOption, error) {
	var (
		options []models.DimensionOption
	)

	if ds.BadRequest {
		return nil, badRequestError
	}

	if ds.DimensionNotFound {
		return nil, dimensionionNotFound
	}

	if ds.InternalError {
		return nil, internalServerError
	}

	option := models.DimensionOption{
		DimensionOptionURL: "/filters/123/dimensions/1_age/options/26",
		Option:             "26",
	}

	options = append(options, option)

	return options, nil
}

func (ds *DataStore) GetFilterDimensionOption(filterID string, name string, option string) error {
	if ds.BadRequest {
		return badRequestError
	}

	if ds.OptionNotFound {
		return optionNotFound
	}

	if ds.InternalError {
		return internalServerError
	}

	return nil
}

func (ds *DataStore) RemoveFilterDimension(string, string) error {
	if ds.InternalError {
		return internalServerError
	}

	if ds.BadRequest {
		return badRequestError
	}

	if ds.Forbidden {
		return forbiddenError
	}

	if ds.NotFound {
		return notFoundError
	}

	return nil
}

func (ds *DataStore) RemoveFilterDimensionOption(filterJobId string, name string, option string) error {
	if ds.InternalError {
		return internalServerError
	}

	if ds.BadRequest {
		return badRequestError
	}

	if ds.Forbidden {
		return forbiddenError
	}

	if ds.DimensionNotFound {
		return dimensionionNotFound
	}

	return nil
}

func (ds *DataStore) UpdateFilter(isAuthenticated bool, filterJob *models.Filter) error {
	if ds.InternalError {
		return internalServerError
	}

	if ds.Unauthorised {
		return unauthorisedError
	}

	if ds.BadRequest {
		return badRequestError
	}

	if ds.Forbidden {
		return forbiddenError
	}

	if ds.NotFound {
		return notFoundError
	}
	return nil
}
