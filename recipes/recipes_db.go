/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	"github.com/ottenwbe/recipes-manager/db"
)

// RecipeDB is the interface that all DB implementations have to expose
type RecipeDB interface {
	db.DB
	Recipes
	Clear()
}
