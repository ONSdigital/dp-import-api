package testapi

import (
	"context"
	"io"
	"net/http"

	mockdatastore "github.com/ONSdigital/dp-import-api/mongo/testmongo"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
)

// List of mocked datastores returning different results
var (
	Dstore              = mockdatastore.DataStorer{}
	DstoreNotFound      = mockdatastore.DataStorer{NotFound: true}
	DstoreInternalError = mockdatastore.DataStorer{InternalError: true}
)

// VerifyAuditorCalls checks the calls to auditor
func VerifyAuditorCalls(callInfo struct {
	Ctx    context.Context
	Action string
	Result string
	Params common.Params
}, a string, r string, p common.Params) {
	So(callInfo.Action, ShouldEqual, a)
	So(callInfo.Result, ShouldEqual, r)
	So(callInfo.Params, ShouldResemble, p)
}

// CreateRequestWithAuth adds authentication to request
func CreateRequestWithAuth(method, URL string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, URL, body)
	ctx := r.Context()
	ctx = common.SetCaller(ctx, "someone@ons.gov.uk")
	r = r.WithContext(ctx)
	return r, err
}

// CreateRequestWithOutAuth builds request without authentication
func CreateRequestWithOutAuth(method, URL string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, URL, body)
	return r, err
}

// NewAuditorMock creates a mocked auditor
func NewAuditorMock() *audit.AuditorServiceMock {
	return &audit.AuditorServiceMock{
		RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
			log.Debug("capturing audit event", nil)
			return nil
		},
	}
}
