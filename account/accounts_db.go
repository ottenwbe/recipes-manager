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

package account

import (
	"context"
	"github.com/ottenwbe/recipes-manager/core"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// MongoAccountService stores and manipulates stored accounts in a DB
type MongoAccountService struct {
	DbClient *core.MongoClient
}

const (
	// EMAIL key
	EMAIL = "name"
	// ID key
	ID = "id"
)

// NewMongoAccountService is created and configured
func NewMongoAccountService(db core.DB) *MongoAccountService {

	accountDB := &MongoAccountService{
		DbClient: db.(*core.MongoClient),
	}

	err := accountDB.createTextIndex()
	if err != nil {
		logrus.Error("Error creating Index", err)
	}

	return accountDB
}

// DeleteAccountByID deletes an account indefinitely
func (db *MongoAccountService) DeleteAccountByID(id AccID) error {
	collection := db.getAccountsCollection()
	_, err := collection.DeleteOne(db.ctx(), bson.M{ID: id})
	if err != nil {
		return err
	}
	return nil
}

// DeleteAccountByName deletes an account indefinitely
func (db *MongoAccountService) DeleteAccountByName(name string) error {
	collection := db.getAccountsCollection()
	_, err := collection.DeleteOne(db.ctx(), bson.M{EMAIL: name})
	if err != nil {
		return err
	}
	return nil
}

// NewAccount stores the account information in a db
func (db *MongoAccountService) NewAccount(name string, t Type) (*Account, error) {
	collection := db.getAccountsCollection()

	acc := NewAccount(name, t)

	_, err := collection.InsertOne(db.ctx(), acc)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

// FindAccount by Name
func (db *MongoAccountService) FindAccount(name string) (*Account, error) {
	collection := db.getAccountsCollection()
	var result Account

	err := collection.FindOne(db.ctx(), bson.M{EMAIL: name}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (db *MongoAccountService) getAccountsCollection() *mongo.Collection {
	return db.DbClient.Client.Database("accounts").Collection("accounts")
}

func (*MongoAccountService) ctx() context.Context {
	return context.Background()
}

func (db *MongoAccountService) createTextIndex() error {

	c := db.getAccountsCollection()

	textIndex := mongo.IndexModel{
		Keys: bsonx.Doc{
			{Key: EMAIL, Value: bsonx.String("text")},
		},
		Options: options.Index().SetUnique(true),
	}

	_, err := c.Indexes().CreateOne(db.ctx(), textIndex)

	return err
}
