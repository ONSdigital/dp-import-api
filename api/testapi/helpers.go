package testapi

import (
	"io"
	"net/http"

	mockdatastore "github.com/ONSdigital/dp-import-api/mongo/testmongo"
	dprequest "github.com/ONSdigital/dp-net/request"
)

// List of mocked datastores returning different results
var (
	Dstore              = mockdatastore.DataStorer{}
	DstoreNotFound      = mockdatastore.DataStorer{NotFound: true}
	DstoreInternalError = mockdatastore.DataStorer{InternalError: true}
)

// CreateRequestWithAuth adds authentication to request
func CreateRequestWithAuth(method, URL string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, URL, body)
	ctx := r.Context()
	ctx = dprequest.SetCaller(ctx, "someone@ons.gov.uk")
	r = r.WithContext(ctx)
	return r, err
}

// CreateRequestWithOutAuth builds request without authentication
func CreateRequestWithOutAuth(method, URL string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, URL, body)
	return r, err
}
