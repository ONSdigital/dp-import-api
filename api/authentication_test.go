package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMiddleWareAuthenticationReturnsForbidden(t *testing.T) {
	t.Parallel()
	Convey("When no access token is provide, not found status code is returned", t, func() {
		auth := NewAuthenticator("123", "internal-token")
		r, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestMiddleWareAuthenticationReturnsUnauthorised(t *testing.T) {
	t.Parallel()
	Convey("When a invalid access token is provide, not found status code is returned", t, func() {
		auth := NewAuthenticator("123", "internal-token")
		r, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
		r.Header.Set("internal-token", "12")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestMiddleWareAuthentication(t *testing.T) {
	t.Parallel()
	Convey("When a valid access token is provide, OK code is returned", t, func() {
		auth := NewAuthenticator("123", "internal-token")
		r, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
		r.Header.Set("internal-token", "123")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func mockHTTPHandler(w http.ResponseWriter, r *http.Request) {

}
