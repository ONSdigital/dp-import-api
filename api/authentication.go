package api

import (
	"errors"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
)

// Authenticator structure which holds the secret key for validating clients. This will be replaced in the future, after the `thin-slices` has been delivered
type Authenticator struct {
	secretKey  string
	headerName string
}

// NewAuthenticator is created
func NewAuthenticator(key, headerName string) Authenticator {
	return Authenticator{secretKey: key, headerName: headerName}
}

// Check wraps a HTTP handle. If authentication fails an error code is returned else the HTTP handler is called
func (a *Authenticator) Check(handle func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(a.headerName)
		if key == "" {
			http.Error(w, notFoundError, http.StatusNotFound)
			log.Error(errors.New("client missing auth token in header"), log.Data{"header": a.headerName})
			return
		}
		if key != a.secretKey {
			http.Error(w, notFoundError, http.StatusNotFound)
			log.Error(errors.New("unauthorised access to API"), log.Data{"header": a.headerName})
			return
		}
		// The request has been authenticated, now run the clients request
		handle(w, r)
	})
}
