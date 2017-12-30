package events

import "github.com/ONSdigital/go-ns/avro"

// ObservationImportCompleteEvent is used to notify that all observations of an instance have been imported.
type ObservationImportComplete struct {
	JobID      string `avro:"job_id"`
	InstanceID string `avro:"instance_id"`
}

var observationImportCompleteSchema = `{
  "type": "record",
  "name": "observations-import-complete",
  "fields": [
    {"name": "job_id", "type": "string"},
    {"name": "instance_id", "type": "string"}
  ]
}`

// ObservationImportCompleteEvent is the Avro schema used when all observations have been imported.
var ObservationImportCompleteSchema = avro.Schema{
	Definition: observationImportCompleteSchema,
}
