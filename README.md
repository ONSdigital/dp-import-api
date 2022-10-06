# DP Import API

## Getting started

### MongoDB

* Run ```brew install mongodb```
* Run ```brew services start mongodb```

### kafka

* Run ```brew install java```
* Run ```brew install kafka```
* Run ```brew services start zookeeper```
* Run ```brew services start kafka```

Scripts for updating and debugging Kafka can be found [here](https://github.com/ONSdigital/dp-data-tools)(dp-data-tools)

## Configuration

| Environment variable         | Default                                                        | Description                                                                                          |
| ---------------------------- |----------------------------------------------------------------|------------------------------------------------------------------------------------------------------|
| BIND_ADDR                    | `:21800`                                                       | The host and port to bind to                                                                         |
| HOST                         | `http://localhost:21800`                                       | The host name used to build URLs                                                                     |
| KAFKA_ADDR                   | `localhost:9092`                                               | A list of kafka brokers (TLS-ready)                                                                  |
| KAFKA_VERSION                | `1.0.2`                                                        | The version of (TLS-ready) Kafka being used                                                          |
| KAFKA_MAX_BYTES              | `200000`                                                       | The max message size for kafka producer                                                              |
| KAFKA_SEC_PROTO              | _unset_                                                        | if set to `TLS`, kafka connections will use TLS ([ref-1])                                            |
| KAFKA_SEC_CLIENT_KEY         | _unset_                                                        | PEM for the client key ([ref-1])                                                                     |
| KAFKA_SEC_CLIENT_CERT        | _unset_                                                        | PEM for the client certificate ([ref-1])                                                             |
| KAFKA_SEC_CA_CERTS           | _unset_                                                        | CA cert chain for the server cert ([ref-1])                                                          |
| KAFKA_SEC_SKIP_VERIFY        | false                                                          | ignores server certificate issues if `true` ([ref-1])                                                |
| DATABAKER_IMPORT_TOPIC       | `data-bake-job-available`                                      | The topic to place messages to data-baker                                                            |
| INPUT_FILE_AVAILABLE_TOPIC   | `input-file-available`                                         | The topic to place V4 files                                                                          |
| MONGODB_BIND_ADDR            | localhost:27017                                                | The MongoDB bind address                                                                             |
| MONGODB_USERNAME             |                                                                | The MongoDB Username                                                                                 |
| MONGODB_PASSWORD             |                                                                | The MongoDB Password                                                                                 |
| MONGODB_DATABASE             | imports                                                        | The MongoDB database                                                                                 |
| MONGODB_COLLECTIONS          | ImportsCollection:imports, ImportsLockCollection:imports_locks | The MongoDB collections                                                                              |
| MONGODB_REPLICA_SET          |                                                                | The name of the MongoDB replica set                                                                  |
| MONGODB_ENABLE_READ_CONCERN  | false                                                          | Switch to use (or not) majority read concern                                                         |
| MONGODB_ENABLE_WRITE_CONCERN | true                                                           | Switch to use (or not) majority write concern                                                        |
| MONGODB_CONNECT_TIMEOUT      | 5s                                                             | The timeout when connecting to MongoDB (`time.Duration` format)                                      |
| MONGODB_QUERY_TIMEOUT        | 15s                                                            | The timeout for querying MongoDB (`time.Duration` format)                                            |
| MONGODB_IS_SSL               | false                                                          | Switch to use (or not) TLS when connecting to mongodb                                                |
| DATASET_API_URL              | `http://localhost:22000`                                       | The URL for the DatasetAPI                                                                           |
| RECIPE_API_URL               | `http://localhost:22300`                                       | The URL for the RecipeAPI                                                                            |
| ZEBEDEE_URL                  | `http://localhost:8082`                                        | The URL Zebedee                                                                                      |
| SERVICE_AUTH_TOKEN           | `0C30662F-6CF6-43B0-A96A-954772267FF5`                         | The token used to identify this service when authenticating                                          |
| HEALTHCHECK_INTERVAL         | `30s`                                                          | The time between calling healthcheck endpoints for check subsystems                                  |
| HEALTHCHECK_CRITICAL_TIMEOUT | `90s`                                                          | The time taken for the health changes from warning state to critical due to subsystem check failures |
| DEFAULT_MAXIMUM_LIMIT        | `1000`                                                         | Default maximum limit for pagination                                                                 |
| DEFAULT_LIMIT                | `20`                                                           | Default limit for pagination                                                                         |
| DEFAULT_OFFSET               | `0`                                                            | Default offset for pagination                                                                        |

[ref-1]:  https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls 'kafka TLS examples documentation'

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2022, [Office for National Statistics](https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
