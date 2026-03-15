/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	"io"
)

// RecipeDB is the interface that all DB implementations have to expose
type RecipeDB interface {
	io.Closer
	Recipes
	Ping() error
	Clear()
}
