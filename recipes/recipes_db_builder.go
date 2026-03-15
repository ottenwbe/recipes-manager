/*
 * MIT License - see LICENSE file for details
 */

package recipes

import "github.com/ottenwbe/recipes-manager/config"

// NewDatabaseClient builds a client to communicate with a database
func NewDatabaseClient() (RecipeDB, error) {
	m := &MongoRecipeDB{}
	addr := config.Config.GetString("recipeDB.host")
	err := m.StartDB(addr)
	return m, err
}
