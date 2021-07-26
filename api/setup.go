package api

import (
	"github.com/ONSdigital/dp-import-api/api/testapi"
	"github.com/ONSdigital/dp-import-api/config"
	mongo "github.com/ONSdigital/dp-import-api/mongo/testmongo"
	"github.com/gorilla/mux"
)

var (
	cfg, _ = config.Get()
)

// SetupAPIWith sets up API with given configs
func SetupAPIWith(overrideDataStore *mongo.DataStorer, overrideServiceMock *testapi.JobServiceMock) *ImportAPI {
	if overrideServiceMock == nil {
		overrideServiceMock = &testapi.JobServiceMock{}
	}
	if overrideDataStore == nil {
		overrideDataStore = &testapi.Dstore
	}

	return Setup(mux.NewRouter(), overrideDataStore, overrideServiceMock, cfg)
}
