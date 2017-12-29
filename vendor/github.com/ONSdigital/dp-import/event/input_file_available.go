package event

import "github.com/ONSdigital/go-ns/avro"

// InputFileAvailable is an event produced when a new input file is available to import.
type InputFileAvailable struct {
	JobID      string `avro:"job_id"`
	InstanceID string `avro:"instance_id"`
	URL        string `avro:"file_url"`
}

var inputFileAvailableSchema = `{
  "type": "record",
  "name": "input-file-available",
  "fields": [
    {"name": "job_id", "type": "string"},
    {"name": "instance_id", "type": "string"},
    {"name": "file_url", "type": "string"}
  ]
}`

// InputFileAvailableSchema provides an Avro schema for the InputFileAvailable event.
var InputFileAvailableSchema = &avro.Schema{
	Definition: inputFileAvailableSchema,
}
