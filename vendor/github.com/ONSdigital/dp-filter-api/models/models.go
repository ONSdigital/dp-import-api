package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

// Filter represents a structure for a filter job
type Filter struct {
	DatasetFilterID  string      `json:"dataset_filter_id"`
	DimensionListURL string      `json:"dimension_list_url,omitempty"`
	Dimensions       []Dimension `json:"dimensions,omitempty"`
	Downloads        Downloads   `json:"downloads,omitempty"`
	Events           Events      `json:"events,omitempty"`
	FilterID         string      `json:"filter_job_id,omitempty"`
	State            string      `json:"state,omitempty"`
}

// Dimension represents an object containing a list of dimension values and the dimension name
type Dimension struct {
	DimensionURL string   `json:"dimension_url,omitempty"`
	Name         string   `json:"name,omitempty"`
	Options      []string `json:"options,omitempty"`
}

// Downloads represents a list of file types possible to download
type Downloads struct {
	CSV  DownloadItem `json:"csv,omitempty"`
	JSON DownloadItem `json:"json,omitempty"`
	XLS  DownloadItem `json:"xls,omitempty"`
}

// DownloadItem represents an object containing information for the download item
type DownloadItem struct {
	Size string `json:"size,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Events represents a list of array objects containing event information against the filter job
type Events struct {
	Error []EventItem `json:"error,omitempty"`
	Info  []EventItem `json:"info,omitempty"`
}

// EventItem represents an event object containing event information
type EventItem struct {
	Message string `json:"message,omitempty"`
	Time    string `json:"time,omitempty"`
	Type    string `json:"type,omitempty"`
}

// AddDimension represents dimension information for storing a list of options for a dimension
type AddDimension struct {
	FilterID string
	Name     string
	Options  []string
}

// AddDimensionOption represents dimension option information for storing
// an individual option for a given filter job dimension
type AddDimensionOption struct {
	FilterID string
	Name     string
	Option   string
}

// DimensionOption represents dimension option information
type DimensionOption struct {
	DimensionOptionURL string `json:"dimension_option_url"`
	Option             string `json:"option"`
}

// Validate checks the content of the filter structure
func (filter *Filter) Validate() error {
	if filter.State == "" {
		filter.State = "created"
	}

	var missingFields []string

	if filter.DatasetFilterID == "" {
		missingFields = append(missingFields, "dataset_filter_id")
	}

	if missingFields != nil {
		return fmt.Errorf("Missing mandatory fields: %v", missingFields)
	}

	return nil
}

// CreateFilter manages the creation of a filter from a reader
func CreateFilter(reader io.Reader) (*Filter, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}
	var filter Filter
	err = json.Unmarshal(bytes, &filter)
	if err != nil {
		return nil, errors.New("Failed to parse json body")
	}

	return &filter, nil
}

// CreateDimensionOptions manages the creation of options for a dimension from a reader
func CreateDimensionOptions(reader io.Reader) ([]string, error) {
	var dimension Dimension

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}

	if string(bytes) == "" {
		return dimension.Options, nil
	}

	err = json.Unmarshal(bytes, &dimension)
	if err != nil {
		return nil, errors.New("Failed to parse json body")
	}

	return dimension.Options, nil
}
