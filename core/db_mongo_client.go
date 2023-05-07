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
	"github.com/ottenwbe/recipes-manager/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"sync"
)

var mongoAddress string

func init() {
	mongoAddress = utils.Config.GetString("recipeDB.host")
}

// MongoClient to connect to the mongo dtabase
type MongoClient struct {
	Client *mongo.Client
	//mtx avoids race conditions while connecting to the database and while closing the connection
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
	} else {
		return errors.New("cannot ping since it is not connected")
	}
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
	log.WithField("addr", mongoAddress).Info("Connecting to DB")
	m.Client, err = mongo.NewClient(options.Client().ApplyURI(mongoAddress))
	if err != nil {
		log.WithError(err).Info("Could not create MongoDB client")
		return
	}
	err = m.Client.Connect(ctx())
	if err != nil {
		log.WithError(err).Info("Could not connect to MongoDB")
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
	defaultContext := context.Background()
	return defaultContext
}
