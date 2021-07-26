package apierrors

import (
	"errors"
	"strconv"
)

// A list of error messages that could be returned by Import API
var (
	ErrFailedToParseJSONBody     = errors.New("failed to parse json body")
	ErrFailedToReadRequestBody   = errors.New("failed to read message body")
	ErrInvalidJob                = errors.New("the provided Job is not valid")
	ErrInvalidQueryParameter     = errors.New("invalid query parameter")
	ErrInvalidPositiveInteger    = errors.New("value is not a positive integer")
	ErrInternalServer            = errors.New("internal error")
	ErrInvalidState              = errors.New("invalid state")
	ErrInvalidUploadedFileObject = errors.New("invalid json object received, alias_name and url are required")
	ErrInvalidInstanceID         = errors.New("the instance id was not found in the provided job")
	ErrJobNotFound               = errors.New("job not found")
	ErrMissingProperties         = errors.New("missing properties to create import job")
	ErrUnauthorised              = errors.New("unauthenticated request")

	NotFoundMap = map[error]bool{
		ErrJobNotFound: true,
	}

	BadRequestMap = map[error]bool{
		ErrFailedToParseJSONBody:     true,
		ErrFailedToReadRequestBody:   true,
		ErrInvalidJob:                true,
		ErrInvalidQueryParameter:     true,
		ErrInvalidPositiveInteger:    true,
		ErrInvalidState:              true,
		ErrInvalidUploadedFileObject: true,
		ErrInvalidInstanceID:         true,
		ErrMissingProperties:         true,
	}
)

// ErrorMaximumLimitReached creates an for the given limit
func ErrorMaximumLimitReached(m int) error {
	return errors.New("the maximum limit has been reached, the limit cannot be more than " + strconv.Itoa(m))
}
