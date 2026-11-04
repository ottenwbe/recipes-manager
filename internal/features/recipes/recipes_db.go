/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	"github.com/ottenwbe/recipes-manager/internal/db"
	"github.com/ottenwbe/recipes-manager/models"
)

// Recipes interface is an abstraction for the provider of a collection of recipes, i.e., a data-base or a cache
type Recipes interface {
	List() []*models.Recipe
	IDs(filterQuery *models.RecipeSearchFilter) models.RecipeList
	Num() int64
	Get(id models.RecipeID) *models.Recipe
	GetByName(name string) (*models.Recipe, error)
	Picture(id models.RecipeID, name string) *models.RecipePicture
	Pictures(id models.RecipeID) map[string]*models.RecipePicture
	Random() *models.Recipe
	Insert(recipe *models.Recipe) error
	Update(id models.RecipeID, recipe *models.Recipe) error
	AddPicture(pic *models.RecipePicture) error
	Remove(id models.RecipeID) error
	RemoveByName(name string) error
}

// RecipeDB is the interface that all DB implementations have to expose
type RecipeDB interface {
	db.DB
	Recipes
	Clear()
}
