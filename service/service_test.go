package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/datastore"
	dsmock "github.com/ONSdigital/dp-import-api/datastore/mock"
	"github.com/ONSdigital/dp-import-api/service"
	"github.com/ONSdigital/dp-import-api/service/mock"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-kafka/kafkatest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
)

var (
	errMongo       = errors.New("MongoDB error")
	errKafka       = errors.New("Kafka producer error")
	errServer      = errors.New("HTTP Server error")
	errHealthcheck = errors.New("healthCheck error")
)

var funcDoGetHealthcheckErr = func(cfg *config.Configuration, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return nil, errHealthcheck
}

var funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.HTTPServer {
	return nil
}

var funcDoGetDataStoreErr = func(cfg *config.Configuration) (datastore.DataStorer, error) {
	return nil, errMongo
}

func TestRun(t *testing.T) {

	Convey("Having a set of mocked dependencies", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcMock := &mock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &mock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return nil
			},
		}

		kafkaMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
		}

		failingServerMock := &mock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return errServer
			},
		}

		funcDoGetHealthcheckOk := func(cfg *config.Configuration, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		funcDoGetFailingHTTPSerer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return failingServerMock
		}

		funcDoGetDataStoreOk := func(cfg *config.Configuration) (datastore.DataStorer, error) {
			return &dsmock.DataStorerMock{}, nil
		}

		funcDoGetKafkaProducer := func(failingTopic string) func(ctx context.Context, kafkaBrokers []string, topic string, envMax int) (kafka.IProducer, error) {
			return func(ctx context.Context, kafkaBrokers []string, topic string, envMax int) (kafka.IProducer, error) {
				if topic == failingTopic {
					return nil, errKafka
				}
				return kafkaMock, nil
			}
		}

		Convey("Given that initialising MongoDB returns an error", func() {
			initMock := &mock.InitialiserMock{
				DoGetMongoDataStoreFunc: funcDoGetDataStoreErr,
				DoGetKafkaProducerFunc:  funcDoGetKafkaProducer(""),
				DoGetHealthCheckFunc:    funcDoGetHealthcheckOk,
				DoGetHTTPServerFunc:     funcDoGetHTTPServer,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			serverWg.Add(1)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds, but the initialisation flag is not set and further initialisations are attempted", func() {
				So(err, ShouldBeNil)
				So(svcList.MongoDataStore, ShouldBeFalse)
				So(svcList.DataBakerProducer, ShouldBeTrue)
				So(svcList.DirectProducer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
			})
		})

		Convey("Given that initialising DataBaker kafka producer returns an error", func() {
			initMock := &mock.InitialiserMock{
				DoGetMongoDataStoreFunc: funcDoGetDataStoreOk,
				DoGetKafkaProducerFunc:  funcDoGetKafkaProducer(cfg.DatabakerImportTopic),
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errKafka)
				So(svcList.MongoDataStore, ShouldBeTrue)
				So(svcList.DataBakerProducer, ShouldBeFalse)
				So(svcList.DirectProducer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising Kafka direct producer returns an error", func() {
			initMock := &mock.InitialiserMock{
				DoGetMongoDataStoreFunc: funcDoGetDataStoreOk,
				DoGetKafkaProducerFunc:  funcDoGetKafkaProducer(cfg.InputFileAvailableTopic),
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errKafka)
				So(svcList.MongoDataStore, ShouldBeTrue)
				So(svcList.DataBakerProducer, ShouldBeTrue)
				So(svcList.DirectProducer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising Helthcheck returns an error", func() {
			initMock := &mock.InitialiserMock{
				DoGetMongoDataStoreFunc: funcDoGetDataStoreOk,
				DoGetKafkaProducerFunc:  funcDoGetKafkaProducer(""),
				DoGetHealthCheckFunc:    funcDoGetHealthcheckErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.MongoDataStore, ShouldBeTrue)
				So(svcList.DataBakerProducer, ShouldBeTrue)
				So(svcList.DirectProducer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {

			errAddheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &mock.HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddheckFail },
				StartFunc:    func(ctx context.Context) {},
			}

			initMock := &mock.InitialiserMock{
				DoGetMongoDataStoreFunc: funcDoGetDataStoreOk,
				DoGetKafkaProducerFunc:  funcDoGetKafkaProducer(""),
				DoGetHealthCheckFunc: func(cfg *config.Configuration, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMockAddFail, nil
				},
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails, but all checks try to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, fmt.Sprintf("unable to register checkers: %s", errAddheckFail.Error()))
				So(svcList.MongoDataStore, ShouldBeTrue)
				So(svcList.DataBakerProducer, ShouldBeTrue)
				So(svcList.DirectProducer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 6)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Kafka Data Baker Producer")
				So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Input File Available Producer")
				So(hcMockAddFail.AddCheckCalls()[2].Name, ShouldResemble, "Zebedee")
				So(hcMockAddFail.AddCheckCalls()[3].Name, ShouldResemble, "Mongo DB")
				So(hcMockAddFail.AddCheckCalls()[4].Name, ShouldResemble, "Dataset API")
				So(hcMockAddFail.AddCheckCalls()[5].Name, ShouldResemble, "Recipe API")
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {
			initMock := &mock.InitialiserMock{
				DoGetMongoDataStoreFunc: funcDoGetDataStoreOk,
				DoGetKafkaProducerFunc:  funcDoGetKafkaProducer(""),
				DoGetHealthCheckFunc:    funcDoGetHealthcheckOk,
				DoGetHTTPServerFunc:     funcDoGetHTTPServer,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			serverWg.Add(1)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.MongoDataStore, ShouldBeTrue)
				So(svcList.DataBakerProducer, ShouldBeTrue)
				So(svcList.DirectProducer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
			})

			Convey("The checkers are registered and the healthcheck and http server started", func() {
				So(len(hcMock.AddCheckCalls()), ShouldEqual, 6)
				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Kafka Data Baker Producer")
				So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Input File Available Producer")
				So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Zebedee")
				So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Mongo DB")
				So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "Dataset API")
				So(hcMock.AddCheckCalls()[5].Name, ShouldResemble, "Recipe API")
				So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, ":21800")
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("Given that all dependencies are successfully initialised but the http server fails", func() {

			initMock := &mock.InitialiserMock{
				DoGetMongoDataStoreFunc: funcDoGetDataStoreOk,
				DoGetKafkaProducerFunc:  funcDoGetKafkaProducer(""),
				DoGetHealthCheckFunc:    funcDoGetHealthcheckOk,
				DoGetHTTPServerFunc:     funcDoGetFailingHTTPSerer,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			serverWg.Add(1)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			Convey("Then the error is returned in the error channel", func() {
				sErr := <-svcErrors
				So(sErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
				So(len(failingServerMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Having a correctly initialised service", t, func(c C) {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false
		serverStopped := false

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &mock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &mock.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
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

		Convey("Closing a service does not close uninitialised dependencies", func() {
			svcList := service.NewServiceList(nil)
			svcList.HealthCheck = true
			svc := service.New(cfg, svcList)
			svc.SetServer(serverMock)
			svc.SetHealthCheck(hcMock)
			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
		})

		fullSvcList := &service.ExternalServiceList{
			MongoDataStore:    true,
			DataBakerProducer: true,
			DirectProducer:    true,
			HealthCheck:       true,
			Init:              nil,
		}

		Convey("Closing the service results in all the initialised dependencies being closed in the expected order", func() {
			svc := service.New(cfg, fullSvcList)
			svc.SetServer(serverMock)
			svc.SetHealthCheck(hcMock)
			svc.SetMongoDataStore(mongoMock)
			svc.SetDataBakerProducer(dataBakerKafkaProducerMock)
			svc.SetInputFileAvailableProducer(inputFileProducerAvailableKafkaProducer)
			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(mongoMock.CloseCalls()), ShouldEqual, 1)
			So(len(dataBakerKafkaProducerMock.CloseCalls()), ShouldEqual, 1)
			So(len(inputFileProducerAvailableKafkaProducer.CloseCalls()), ShouldEqual, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			failingserverMock := &mock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			svc := service.New(cfg, fullSvcList)
			svc.SetServer(failingserverMock)
			svc.SetHealthCheck(hcMock)
			svc.SetMongoDataStore(mongoMock)
			svc.SetDataBakerProducer(dataBakerKafkaProducerMock)
			svc.SetInputFileAvailableProducer(inputFileProducerAvailableKafkaProducer)
			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "failed to shutdown gracefully")
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(failingserverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(mongoMock.CloseCalls()), ShouldEqual, 1)
			So(len(dataBakerKafkaProducerMock.CloseCalls()), ShouldEqual, 1)
			So(len(inputFileProducerAvailableKafkaProducer.CloseCalls()), ShouldEqual, 1)
		})
	})
}
