/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	"github.com/ottenwbe/recipes-manager/core"
)

// RecipeDB is the interface that all DB implementations have to expose
type RecipeDB interface {
	core.DB
	Recipes
	Clear()
}
