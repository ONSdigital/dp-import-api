module github.com/ONSdigital/dp-import-api

go 1.19

// to avoid 'sonatype-2021-4899' non-CVE Vulnerability
exclude github.com/gorilla/sessions v1.2.1

// to avoid the following vulnerabilities:
//     - CVE-2022-29153 # pkg:golang/github.com/hashicorp/consul/api@v1.1.0
//     - sonatype-2021-1401 # pkg:golang/github.com/miekg/dns@v1.0.14
replace github.com/spf13/cobra => github.com/spf13/cobra v1.4.0

require (
	github.com/ONSdigital/dp-api-clients-go v1.43.0
	github.com/ONSdigital/dp-api-clients-go/v2 v2.185.0
	github.com/ONSdigital/dp-healthcheck v1.3.0
	github.com/ONSdigital/dp-import v1.3.1
	github.com/ONSdigital/dp-kafka/v2 v2.7.3
	github.com/ONSdigital/dp-mongodb/v3 v3.3.0
	github.com/ONSdigital/dp-net v1.5.0
	github.com/ONSdigital/log.go/v2 v2.2.0
	github.com/gorilla/mux v1.8.0
	github.com/justinas/alice v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/smartystreets/goconvey v1.7.2
	go.mongodb.org/mongo-driver v1.10.3
)

require (
	github.com/ONSdigital/dp-net/v2 v2.4.0 // indirect
	github.com/Shopify/sarama v1.37.2 // indirect
	github.com/aws/aws-sdk-go v1.44.112 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-avro/avro v0.0.0-20171219232920-444163702c11 // indirect
	github.com/go-test/deep v1.0.7 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/smartystreets/assertions v1.13.0 // indirect
	github.com/square/mongo-lock v0.0.0-20220601164918-701ecf357cd7 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	golang.org/x/crypto v0.0.0-20221005025214-4161e89ecf1b // indirect
	golang.org/x/net v0.0.0-20221004154528-8021a29435af // indirect
	golang.org/x/sync v0.0.0-20220929204114-8fcdb60fdcc0 // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	golang.org/x/text v0.3.8 // indirect
)
