// SPDX-License-Identifier: MIT

package core

import (
	"io"

	"github.com/ottenwbe/recipes-manager/config"
)

// RecipeDBHostCfg key for configuration
const RecipeDBHostCfg = "recipeDB.host"

func init() {
	config.Config.SetDefault(RecipeDBHostCfg, "mongodb://127.0.0.1:27017")
}

// DB is the interface that all DB implementations have to expose
type DB interface {
	io.Closer
	Ping() error
}

// NewDatabaseClient builds a client to communicate with a database
func NewDatabaseClient(addr string) (DB, error) {
	m := &MongoClient{}
	err := m.StartDB(addr)
	return m, err
}
