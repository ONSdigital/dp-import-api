package api_errors

import (
	"errors"
)

var JobNotFoundError = errors.New("No job found")

var ForbiddenOperation = errors.New("Forbidden operation")

var DimensionNameNotFoundError = errors.New("No dimension name found")
