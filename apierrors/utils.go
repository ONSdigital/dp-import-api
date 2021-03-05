package apierrors

import (
	"errors"
)

// A list of error messages that could be returned by Import API
var (
	ErrFailedToParseJSONBody     = errors.New("failed to parse json body")
	ErrFailedToReadRequestBody   = errors.New("failed to read message body")
	ErrInvalidJob                = errors.New("the provided Job is not valid")
	ErrInvalidQueryParameter     = errors.New("invalid query parameter")
	ErrInternalServer            = errors.New("internal error")
	ErrInvalidState              = errors.New("invalid state")
	ErrInvalidUploadedFileObject = errors.New("invalid json object received, alias_name and url are required")
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
		ErrInvalidState:              true,
		ErrInvalidUploadedFileObject: true,
		ErrMissingProperties:         true,
	}
)
