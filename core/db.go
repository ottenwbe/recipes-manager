/*
 * MIT License - see LICENSE file for details
 */

package core

import (
	"io"
)

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
