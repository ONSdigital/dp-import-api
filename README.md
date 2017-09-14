DP Import API
==============

### Getting started

#### MongoDB
* Run ```brew install mongodb```
* Run ```brew services start mongodb```

#### kafka
* Run ```brew install java```
* Run ```brew install kafka```
* Run ```brew services start zookeeper```
* Run ```brew services start kafka```

### Configuration

| Environment variable | Default                                   | Description
| -------------------- | ----------------------------------------- | -----------
| BIND_ADDR            | :21800                                    | The host and port to bind to
| POSTGRES_URL         | user=dp dbname=ImportJobs sslmode=disable | URL to a Postgres services
| KAFKA_ADDR           | localhost:9092                            | A list of kafka brokers
| KAFKA_MAX_BYTES      | 200000                                    | The max message size for kafka producer
| DATABAKER_IMPORT_TOPIC | data-bake-job-available                 | The topic to place messages to data-baker
| INPUT_FILE_AVAILABLE_TOPIC  | input-file-available               | The topic to place V4 files
| HOST                 |  "http://localhost:21800"                 | The host name used to build URLs
| SECRET_KEY           | "FD0108EA-825D-411C-9B1D-41EF7727F465"    | A key used by the API
| MONGODB_IMPORTS_ADDR       | "localhost:27017"                       | Address of MongoDB
| MONGODB_IMPORTS_DATABASE   | "imports"                               | The mongodb database to store imports
| MONGODB_IMPORTS_COLLECTION | "imports"                               | The mongodb collection to store imports
| DATASET_API_URL            | "http://localhost:22000"                | The URL for the DatasetAPI
| DATASET_AUTH_TOKEN         | "FD0108EA-825D-411C-9B1D-41EF7727F465"  | The Auth Token for the DatasetAPI

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
