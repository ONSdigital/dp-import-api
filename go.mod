module github.com/ONSdigital/dp-import-api

go 1.17

// workaround for insecurity
// https://ossindex.sonatype.org/vulnerability/bba60acb-c7b5-4621-af69-f4085a8301d0?component-type=golang&component-name=github.com%2Fcoreos%2Fetcd&utm_source=nancy-client&utm_medium=integration&utm_content=1.0.22
// CVE-2020-15136
// CVE-2020-15115
// CVE-2020-15114
replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible

require (
	github.com/ONSdigital/dp-api-clients-go v1.43.0
	github.com/ONSdigital/dp-api-clients-go/v2 v2.1.7-beta
	github.com/ONSdigital/dp-healthcheck v1.1.3
	github.com/ONSdigital/dp-import v1.2.1
	github.com/ONSdigital/dp-kafka/v2 v2.4.1
	github.com/ONSdigital/dp-mongodb v1.7.0
	github.com/ONSdigital/dp-net v1.2.0
	github.com/ONSdigital/log.go/v2 v2.0.9
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/gorilla/mux v1.8.0
	github.com/justinas/alice v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/smartystreets/goconvey v1.6.4
)

require (
	github.com/Shopify/sarama v1.29.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fatih/color v1.12.0 // indirect
	github.com/go-avro/avro v0.0.0-20171219232920-444163702c11 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20210803090616-8f023c250c89 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20210113012101-fb4e108d2519 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/compress v1.13.5 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/square/mongo-lock v0.0.0-20191001051310-282c90e422d0 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/net v0.0.0-20210903162142-ad29c8ab022f // indirect
	golang.org/x/sys v0.0.0-20210903071746-97244b99971b // indirect
)
