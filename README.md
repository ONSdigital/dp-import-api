DP Import API
==============


### Getting started

#### Postgres
* Run ```brew install postgres```
* Run ```brew services start postgres```
* Run ```createuser dp -d -w```
* Run ```createdb --owner dp ImportJobs```
* Run ```psql -U dp ImportJobs -f scripts/InitDatabase.sql```

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

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.