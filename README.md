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
| DB_URL               | user=dp dbname=ImportJobs sslmode=disable | URL to a Postgres services

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.