// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package testjob

import (
	"github.com/ONSdigital/dp-import-api/models"
	"sync"
)

var (
	lockRecipeAPIMockGetRecipe sync.RWMutex
)

// RecipeAPIMock is a mock implementation of RecipeAPI.
//
//     func TestSomethingThatUsesRecipeAPI(t *testing.T) {
//
//         // make and configure a mocked RecipeAPI
//         mockedRecipeAPI := &RecipeAPIMock{
//             GetRecipeFunc: func(url string) (*models.Recipe, error) {
// 	               panic("TODO: mock out the GetRecipe method")
//             },
//         }
//
//         // TODO: use mockedRecipeAPI in code that requires RecipeAPI
//         //       and then make assertions.
//
//     }
type RecipeAPIMock struct {
	// GetRecipeFunc mocks the GetRecipe method.
	GetRecipeFunc func(url string) (*models.Recipe, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetRecipe holds details about calls to the GetRecipe method.
		GetRecipe []struct {
			// Url is the url argument value.
			Url string
		}
	}
}

// GetRecipe calls GetRecipeFunc.
func (mock *RecipeAPIMock) GetRecipe(url string) (*models.Recipe, error) {
	if mock.GetRecipeFunc == nil {
		panic("moq: RecipeAPIMock.GetRecipeFunc is nil but RecipeAPI.GetRecipe was just called")
	}
	callInfo := struct {
		Url string
	}{
		Url: url,
	}
	lockRecipeAPIMockGetRecipe.Lock()
	mock.calls.GetRecipe = append(mock.calls.GetRecipe, callInfo)
	lockRecipeAPIMockGetRecipe.Unlock()
	return mock.GetRecipeFunc(url)
}

// GetRecipeCalls gets all the calls that were made to GetRecipe.
// Check the length with:
//     len(mockedRecipeAPI.GetRecipeCalls())
func (mock *RecipeAPIMock) GetRecipeCalls() []struct {
	Url string
} {
	var calls []struct {
		Url string
	}
	lockRecipeAPIMockGetRecipe.RLock()
	calls = mock.calls.GetRecipe
	lockRecipeAPIMockGetRecipe.RUnlock()
	return calls
}
