/*
 * MIT License - see LICENSE file for details
 */

package core

import (
	"io"

	"github.com/ottenwbe/recipes-manager/config"
)

// DB is the interface that all DB implementations have to expose
type DB interface {
	io.Closer
	Ping() error
}

// NewDatabaseClient builds a client to communicate with a database
func NewDatabaseClient() (DB, error) {
	m := &MongoClient{}
	addr := config.Config.GetString("recipeDB.host")
	err := m.StartDB(addr)
	return m, err
}
