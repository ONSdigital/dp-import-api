package dataset

import "github.com/ONSdigital/dp-import-api/models"

type DatasetAPIer interface {
	CreateInstance(jobID, jobURL string, recipeInstance *models.RecipeInstance) (*models.Instance, error)
	GetURL() string
	UpdateInstanceState(instanceID, newState string) error
}
