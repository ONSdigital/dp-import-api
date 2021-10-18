package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/datastore"
	dsmock "github.com/ONSdigital/dp-import-api/datastore/mock"
	"github.com/ONSdigital/dp-import-api/service/mock"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "1599210455"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
)

var (
	errMongo       = errors.New("MongoDB error")
	errKafka       = errors.New("Kafka producer error")
	errServer      = errors.New("HTTP Server error")
	errHealthcheck = errors.New("healthCheck error")
)

func TestNew(t *testing.T) {
	Convey("New returns a new uninitialised service", t, func() {
		So(New(), ShouldResemble, &Service{})
	})
}

func TestInit(t *testing.T) {

	Convey("Given a set of mocked dependencies", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		datastoreMock := &dsmock.DataStorerMock{}
		getMongoDataStore = func(ctx context.Context, cfg *config.Configuration) (datastore.DataStorer, error) {
			return datastoreMock, nil
		}

		kafkaMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
		}
		getKafkaProducer = func(ctx context.Context, cfg *config.KafkaConfig, topic string) (kafka.IProducer, error) {
			return kafkaMock, nil
		}
		getKafkaProducer = func(ctx context.Context, cfg *config.KafkaConfig, topic string) (kafka.IProducer, error) {
			return kafkaMock, nil
		}

		hcMock := &mock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		}
		getHealthCheck = func(version healthcheck.VersionInfo, criticalTimeout, interval time.Duration) HealthChecker {
			return hcMock
		}

		serverMock := &mock.HTTPServerMock{}
		getHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
			return serverMock
		}

		svc := &Service{}

		Convey("When initialising MongoDB returns an error", func() {
			getMongoDataStore = func(ctx context.Context, cfg *config.Configuration) (datastore.DataStorer, error) {
				return nil, errMongo
			}

			Convey("Then service Init succeeds, mongoDataStore dependency is not set and further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.mongoDataStore, ShouldBeNil)
				So(svc.dataBakerProducer, ShouldResemble, kafkaMock)
				So(svc.inputFileAvailableProducer, ShouldResemble, kafkaMock)
				So(svc.cantabularDatasetInstanceStartedProducer, ShouldResemble, kafkaMock)
				So(svc.healthCheck, ShouldResemble, hcMock)
				So(svc.server, ShouldResemble, serverMock)

				Convey("But all checks try to register", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 7)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Kafka Data Baker Producer")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Input File Available Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Kafka Cantabular Dataset Instance Started Producer")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Zebedee")
					So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "Mongo DB")
					So(hcMock.AddCheckCalls()[5].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[6].Name, ShouldResemble, "Recipe API")
				})
			})
		})

		Convey("When initialising DataBaker kafka producer returns an error", func() {
			getKafkaProducer = func(ctx context.Context, kafkaCfg *config.KafkaConfig, topic string) (kafka.IProducer, error) {
				if topic == kafkaCfg.DatabakerImportTopic {
					return nil, errKafka
				}
				return kafkaMock, nil
			}

			Convey("Then service Init fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldResemble, errKafka)
				So(svc.mongoDataStore, ShouldResemble, datastoreMock)
				So(svc.dataBakerProducer, ShouldBeNil)
				So(svc.inputFileAvailableProducer, ShouldBeNil)
				So(svc.cantabularDatasetInstanceStartedProducer, ShouldBeNil)
				So(svc.healthCheck, ShouldBeNil)
				So(svc.server, ShouldBeNil)
			})
		})

		Convey("When initialising Kafka direct producer returns an error", func() {
			getKafkaProducer = func(ctx context.Context, kafkaCfg *config.KafkaConfig, topic string) (kafka.IProducer, error) {
				if topic == kafkaCfg.InputFileAvailableTopic {
					return nil, errKafka
				}
				return kafkaMock, nil
			}

			Convey("Then service Init fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldResemble, errKafka)
				So(svc.mongoDataStore, ShouldResemble, datastoreMock)
				So(svc.dataBakerProducer, ShouldResemble, kafkaMock)
				So(svc.inputFileAvailableProducer, ShouldBeNil)
				So(svc.cantabularDatasetInstanceStartedProducer, ShouldBeNil)
				So(svc.healthCheck, ShouldBeNil)
				So(svc.server, ShouldBeNil)
			})
		})

		Convey("When initialising Kafka cantabular producer returns an error", func() {
			getKafkaProducer = func(ctx context.Context, kafkaCfg *config.KafkaConfig, topic string) (kafka.IProducer, error) {
				if topic == kafkaCfg.CantabularDatasetInstanceStartedTopic {
					return nil, errKafka
				}
				return kafkaMock, nil
			}

			Convey("Then service Init fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldResemble, errKafka)
				So(svc.mongoDataStore, ShouldResemble, datastoreMock)
				So(svc.dataBakerProducer, ShouldResemble, kafkaMock)
				So(svc.inputFileAvailableProducer, ShouldResemble, kafkaMock)
				So(svc.cantabularDatasetInstanceStartedProducer, ShouldBeNil)
				So(svc.cantabularDatasetInstanceStartedProducer, ShouldBeNil)
				So(svc.healthCheck, ShouldBeNil)
				So(svc.server, ShouldBeNil)
			})
		})

		Convey("When healthcheck versionInfo cannot be created due to a wrong build time", func() {
			wrongBuildTime := "wrongFormat"

			Convey("Then service Init fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, wrongBuildTime, testGitCommit, testVersion)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "failed to parse build time")
				So(svc.mongoDataStore, ShouldResemble, datastoreMock)
				So(svc.dataBakerProducer, ShouldResemble, kafkaMock)
				So(svc.inputFileAvailableProducer, ShouldResemble, kafkaMock)
				So(svc.cantabularDatasetInstanceStartedProducer, ShouldResemble, kafkaMock)
				So(svc.healthCheck, ShouldBeNil)
				So(svc.server, ShouldBeNil)
			})
		})

		Convey("When Checkers cannot be registered", func() {
			hcMock.AddCheckFunc = func(name string, checker healthcheck.Checker) error { return errHealthcheck }

			Convey("Then service Init fails with the expected error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "unable to register checkers: Error(s) registering checkers for healthcheck")
				So(svc.mongoDataStore, ShouldResemble, datastoreMock)
				So(svc.dataBakerProducer, ShouldResemble, kafkaMock)
				So(svc.inputFileAvailableProducer, ShouldResemble, kafkaMock)
				So(svc.cantabularDatasetInstanceStartedProducer, ShouldResemble, kafkaMock)
				So(svc.healthCheck, ShouldResemble, hcMock)
				So(svc.server, ShouldBeNil)

				Convey("But all checks try to register", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 7)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Kafka Data Baker Producer")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Input File Available Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Kafka Cantabular Dataset Instance Started Producer")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Zebedee")
					So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "Mongo DB")
					So(hcMock.AddCheckCalls()[5].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[6].Name, ShouldResemble, "Recipe API")
				})
			})
		})

		Convey("When all dependencies are successfully initialised", func() {

			Convey("Then service Init succeeds and all the flags are set", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.mongoDataStore, ShouldResemble, datastoreMock)
				So(svc.dataBakerProducer, ShouldResemble, kafkaMock)
				So(svc.inputFileAvailableProducer, ShouldResemble, kafkaMock)
				So(svc.healthCheck, ShouldResemble, hcMock)
				So(svc.server, ShouldResemble, serverMock)

				Convey("And all checks are registered", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 7)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Kafka Data Baker Producer")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Input File Available Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Kafka Cantabular Dataset Instance Started Producer")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Zebedee")
					So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "Mongo DB")
					So(hcMock.AddCheckCalls()[5].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[6].Name, ShouldResemble, "Recipe API")
				})
			})
		})
	})
}

func TestStart(t *testing.T) {

	Convey("Given a correctly initialised Service with mocked dependencies", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		kafkaProducerDirect := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
		}

		kafkaProducerBaker := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
		}

		kafkaProducerCantabular := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
		}

		hcMock := &mock.HealthCheckerMock{
			StartFunc: func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &mock.HTTPServerMock{}

		svc := &Service{
			cfg:                                      cfg,
			dataBakerProducer:                        kafkaProducerBaker,
			inputFileAvailableProducer:               kafkaProducerDirect,
			cantabularDatasetInstanceStartedProducer: kafkaProducerCantabular,
			healthCheck:                              hcMock,
			server:                                   serverMock,
		}

		Convey("When a service with a successful HTTP server is started", func() {
			serverMock.ListenAndServeFunc = func() error {
				serverWg.Done()
				return nil
			}
			serverWg.Add(1)
			svc.Start(ctx, make(chan error, 1))

			Convey("Then healthcheck is started and HTTP server starts listening", func() {
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a service with a failing HTTP server is started", func() {
			serverMock.ListenAndServeFunc = func() error {
				serverWg.Done()
				return errServer
			}
			errChan := make(chan error, 1)
			serverWg.Add(1)
			svc.Start(ctx, errChan)

			Convey("Then HTTP server errors are reported to the provided errors channel", func() {
				rxErr := <-errChan
				So(rxErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Given a correctly initialised service with mocked dependencies", t, func(c C) {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false
		serverStopped := false

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &mock.HealthCheckerMock{
			StopFunc: func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &mock.HTTPServerMock{
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server was stopped before healthcheck")
				}
				serverStopped = true
				return nil
			},
		}

		funcClose := func(ctx context.Context) error {
			if !hcStopped {
				return errors.New("Dependency was closed before healthcheck")
			}
			if !serverStopped {
				return errors.New("Dependency was closed before http server")
			}
			return nil
		}

		// mongoDB will fail if healthcheck or http server are not stopped
		mongoMock := &dsmock.DataStorerMock{
			CloseFunc: funcClose,
		}

		// dataBakerKafkaProducerMock producer will fail if healthcheck or http server are not stopped
		dataBakerKafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
			CloseFunc: funcClose,
		}

		// inputFileProducerAvailableKafkaProducer producer will fail if healthcheck or http server are not stopped
		inputFileProducerAvailableKafkaProducer := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
			CloseFunc: funcClose,
		}

		cantabularKafkaProducer := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
			CloseFunc: funcClose,
		}

		Convey("Closing a service does not close uninitialised dependencies", func() {
			svc := Service{
				cfg: cfg,
			}
			err := svc.Close(context.Background())
			So(err, ShouldBeNil)
		})

		svc := Service{
			cfg:                                      cfg,
			healthCheck:                              hcMock,
			mongoDataStore:                           mongoMock,
			dataBakerProducer:                        dataBakerKafkaProducerMock,
			inputFileAvailableProducer:               inputFileProducerAvailableKafkaProducer,
			cantabularDatasetInstanceStartedProducer: cantabularKafkaProducer,
			server:                                   serverMock,
		}

		Convey("Closing the service results in all the initialised dependencies being closed in the expected order", func() {
			err := svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(mongoMock.CloseCalls()), ShouldEqual, 1)
			So(len(dataBakerKafkaProducerMock.CloseCalls()), ShouldEqual, 1)
			So(len(inputFileProducerAvailableKafkaProducer.CloseCalls()), ShouldEqual, 1)
			So(len(cantabularKafkaProducer.CloseCalls()), ShouldEqual, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			serverMock.ShutdownFunc = func(ctx context.Context) error {
				return errors.New("Failed to stop http server")
			}

			err := svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "failed to shutdown gracefully")
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(mongoMock.CloseCalls()), ShouldEqual, 1)
			So(len(dataBakerKafkaProducerMock.CloseCalls()), ShouldEqual, 1)
			So(len(inputFileProducerAvailableKafkaProducer.CloseCalls()), ShouldEqual, 1)
			So(len(cantabularKafkaProducer.CloseCalls()), ShouldEqual, 1)
		})

		Convey("When a dependency takes more time to close than the graceful shutdown timeout", func() {
			cfg.GracefulShutdownTimeout = 1 * time.Millisecond
			serverMock.ShutdownFunc = func(ctx context.Context) error {
				time.Sleep(20 * time.Millisecond)
				return nil
			}

			Convey("Then closing the service fails with context.DeadlineExceeded error and no further dependencies are attempted to close", func() {
				err = svc.Close(context.Background())
				So(err, ShouldResemble, context.DeadlineExceeded)
				So(len(hcMock.StopCalls()), ShouldEqual, 1)
				So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
				So(len(mongoMock.CloseCalls()), ShouldEqual, 0)
				So(len(dataBakerKafkaProducerMock.CloseCalls()), ShouldEqual, 0)
				So(len(inputFileProducerAvailableKafkaProducer.CloseCalls()), ShouldEqual, 0)
				So(len(cantabularKafkaProducer.CloseCalls()), ShouldEqual, 0)
			})
		})
	})
}
