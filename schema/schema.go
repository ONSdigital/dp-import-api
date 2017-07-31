package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var DataBakerEvent = `{
  "type": "record",
  "name": "publish-dataset",
  "fields": [
  { "name": "job_id", "type": "string" }
  ]
}`

var ImportV4FileEvent = `{
  "type": "record",
  "name": "input-file-available",
  "fields": [
    {"name": "file_url", "type": "string"},
    {"name": "instance_id", "type": "string"}
  ]
}`

var DataBaker *avro.Schema = &avro.Schema{
	Definition: DataBakerEvent,
}

var ImportV4File *avro.Schema = &avro.Schema{
	Definition: ImportV4FileEvent,
}
