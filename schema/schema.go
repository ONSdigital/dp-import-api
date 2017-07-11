package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var PublishDatasetEvent = `{
  "type": "record",
  "name": "publish-dataset",
  "fields": [
  { "name": "recipe", "type": "string" },
  { "name": "instance_ids", "type": { "type": "array", "items": "string"}},
  { "name": "files", "type": { "type": "array", "items": {
    "name": "file", "type": "record", "fields": [
     { "name": "alias-name", "type": "string"},
     { "name": "url", "type": "string"}
]
}
}
}
]
}`

var PublishDataset *avro.Schema = &avro.Schema{
	Definition: PublishDatasetEvent,
}
