/*
 * MIT License
 *
 * Copyright (c) 2023 Beate Ottenwälder
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package core

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ottenwbe/recipes-manager/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// getMongoAddress returns the configured database host.
// Retrieving it dynamically avoids issues with package initialization order.
func getMongoAddress() string {
	return utils.Config.GetString("recipeDB.host")
}

// MongoClient to connect to the mongo database
type MongoClient struct {
	Client *mongo.Client
	// mtx avoids race conditions while connecting to the database and while closing the connection
	mtx sync.Mutex
}

// Close the connection to the database
func (m *MongoClient) Close() error {
	return m.StopDB()
}

// Ping MongoDB
func (m *MongoClient) Ping() error {

	if m.Client != nil {
		return m.Client.Ping(ctx(), readpref.Primary())
	}
	return errors.New("cannot ping since it is not connected")
}

// StartDB initializes the database connection
func (m *MongoClient) StartDB() error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.Client == nil {

		err := m.connectToDB()
		if err != nil {
			log.WithError(err).Error("Database is not connected")
			return errors.New("database is not connected")
		}
	} else {
		return errors.New("database is already running")
	}

	return nil
}

// StopDB closes the connection to the db
func (m *MongoClient) StopDB() (err error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.Client != nil {
		err = m.Client.Disconnect(ctx())
	}
	m.Client = nil

	return
}

func (m *MongoClient) connectToDB() (err error) {
	addr := getMongoAddress()
	if addr == "" {
		log.Warn("MongoDB address is empty, check configuration")
	}

	log.WithField("addr", addr).Info("Connecting to DB")
	m.Client, err = mongo.Connect(options.Client().ApplyURI(addr))
	if err != nil {
		log.WithError(err).Info("Could not create and connect MongoDB client")
		return
	}
	err = m.Ping()
	if err != nil {
		log.WithError(err).Info("Could not ping MongoDB")
		return
	}

	return
}

func ctx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return ctx
}
