/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	"errors"
	"fmt"

	"github.com/ottenwbe/recipes-manager/core"
)

// NewRecipeDB builds a client to communicate with a database from a generic db client
func NewRecipeDB(db core.DB) (RecipeDB, error) {
	mongoClient, ok := db.(*core.MongoClient)
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
