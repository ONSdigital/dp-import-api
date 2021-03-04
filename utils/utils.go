package utils

import (
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"strconv"
)

// ValidatePositiveInt obtains the positive int value of query var defined by the provided varKey
func ValidatePositiveInt(parameter string) (val int, err error) {

	val, err = strconv.Atoi(parameter)
	if err != nil {
		return -1, errs.ErrInvalidQueryParameter
	}
	if val < 0 {
		return -1, errs.ErrInvalidQueryParameter
	}
	return val, nil
}
