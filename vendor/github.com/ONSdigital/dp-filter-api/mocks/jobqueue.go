package mocks

import (
	"fmt"

	"github.com/ONSdigital/dp-filter-api/models"
)

type FilterJob struct {
	ReturnError bool
}

type MessageData struct {
	FilterJobID string
}

func (fj *FilterJob) Queue(filter *models.Filter) error {
	if fj.ReturnError {
		return fmt.Errorf("No message produced for filter job")
	}
	return nil
}
