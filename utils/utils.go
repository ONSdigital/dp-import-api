package utils

import (
	"errors"
	"os"
)

var JobNotFoundError = errors.New("No job found")

func GetEnvVariable(name, defaultValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	}
	return defaultValue
}
