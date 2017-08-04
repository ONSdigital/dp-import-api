package api_errors

import (
	"errors"
)

var JobNotFoundError = errors.New("No job found")

var DimensionNameNotFoundError = errors.New("No dimension name found")
