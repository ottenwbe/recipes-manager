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
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

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
		c, cancel := ctx()
		defer cancel()
		return m.Client.Ping(c, readpref.Primary())
	}
	return errors.New("cannot ping since it is not connected")
}

// StartDB initializes the database connection
func (m *MongoClient) StartDB(addr string) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.Client == nil {

		err := m.connectToDB(addr)
		if err != nil {
			return fmt.Errorf("database is not connected: %w", err)
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
		c, cancel := ctx()
		defer cancel()
		err = m.Client.Disconnect(c)
	}
	m.Client = nil

	return
}

func (m *MongoClient) connectToDB(addr string) (err error) {
	if addr == "" {
		log.Warn("MongoDB address is empty, check configuration")
	}

	log.WithField("addr", addr).Info("Connecting to DB")
	m.Client, err = mongo.Connect(options.Client().ApplyURI(addr))
	if err != nil {
		log.WithError(err).Error("Could not create and connect MongoDB client")
		return err
	}
	err = m.Ping()
	if err != nil {
		log.WithError(err).Error("Could not ping MongoDB")
		return err
	}

	return
}

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
