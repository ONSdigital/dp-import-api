package recipe_test

import (
	"bytes"
	"context"
	"github.com/ONSdigital/dp-import-api/recipe"
	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetRecipe(t *testing.T) {

	Convey("Given a 404 response from the recipe API", t, func() {

		testCtx := context.Background()

		mockHttpClient := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return Response(make([]byte, 0), 404), nil
			},
		}

		recipeAPI := recipe.API{
			URL:    "",
			Client: mockHttpClient,
		}

		Convey("When the GetInstance method is called", func() {

			recipe, err := recipeAPI.GetRecipe(testCtx, "RecipeID")

			Convey("Then the expected response is returned with no error", func() {
				So(recipe, ShouldBeNil)
				So(err.Error(), ShouldEqual, "bad response")
			})
		})
	})
}

func Response(body []byte, statusCode int) *http.Response {
	reader := bytes.NewBuffer(body)
	readCloser := ioutil.NopCloser(reader)

	return &http.Response{
		StatusCode: statusCode,
		Body:       readCloser,
	}
}
