/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	"errors"
	"fmt"

	"github.com/ottenwbe/recipes-manager/db"
)

// NewRecipeDB builds a client to communicate with a generic db client from a generic db client
func NewRecipeDB(database db.DB) (RecipeDB, error) {
	mongoClient, ok := database.(*db.MongoClient)
	if !ok {
		return nil, errors.New("provided database is not a mongo client")
	}

	var err error
	m := &MongoRecipeDB{mongoClient: mongoClient}
	if err := m.ensureIndexes(); err != nil {
		return nil, fmt.Errorf("could not ensure indexes: %w", err)
	}
	return m, err
}
