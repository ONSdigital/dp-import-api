package api

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"

)

func TestMiddleWareAuthenticationReturnsForbidden(t *testing.T) {
	t.Parallel()
	Convey("When no access token is provide, forbidden status code is returned", t, func() {
		auth := NewAuthenticator("123", "internal-token")
		r, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.MiddleWareAuthentication(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)
	})
}

func TestMiddleWareAuthenticationReturnsUnauthorised(t *testing.T) {
	t.Parallel()
	Convey("When a invalid access token is provide, unauthorised status code is returned", t, func() {
		auth := NewAuthenticator("123", "internal-token")
		r, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
		r.Header.Set("internal-token","12")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.MiddleWareAuthentication(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})
}

func TestMiddleWareAuthentication(t *testing.T) {
	t.Parallel()
	Convey("When a valid access token is provide, OK code is returned", t, func() {
		auth := NewAuthenticator("123", "internal-token")
		r, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
		r.Header.Set("internal-token","123")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.MiddleWareAuthentication(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}


func TestMiddleWareAuthenticationWithValue(t *testing.T) {
	t.Parallel()
	Convey("When a valid access token is provide, true is passed to a http handler", t, func() {
		auth := NewAuthenticator("123", "internal-token")
		r, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
		r.Header.Set("internal-token","123")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		var isRequestAuthenticated bool
		auth.MiddleWareAuthenticationWithValue(func(w http.ResponseWriter, r *http.Request, isAuth bool) {
			isRequestAuthenticated = isAuth
		}).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(isRequestAuthenticated, ShouldEqual, true)
	})
}

func mockHTTPHandler(w http.ResponseWriter, r *http.Request) {

}
