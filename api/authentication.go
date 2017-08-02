package api

import (
	"net/http"
	"github.com/ONSdigital/go-ns/log"
	"errors"
)

type Authenticator struct {
	secretKey string
	headerName string
}

func NewAuthenticator(key, headerName string) Authenticator {
	return Authenticator{ secretKey: key, headerName: headerName}
}

func (a Authenticator) MiddleWareAuthentication(handle func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(a.headerName)
		if key == "" {
			http.Error(w, "No authentication cookie provided", http.StatusForbidden)
			log.Error(errors.New("Client missing token"), log.Data{"header":a.headerName})
			return
		}
		if key != a.secretKey {
			http.Error(w, "Unauthorised access to API", http.StatusUnauthorized)
			log.Error(errors.New("Unauthorised access to API"), log.Data{"header":a.headerName})
			return
		}
		// The request has been authenticated, now run the clients request
		handle(w, r)
	})
}
