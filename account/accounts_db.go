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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoAccountDB struct {
	Db *core.MongoClient
}

func NewMongoAccountClient(db core.DB) *MongoAccountDB {
	return &MongoAccountDB{
		Db: db.(*core.MongoClient),
	}
}

func (db *MongoAccountDB) DeleteAccount(acc *Account) error {
	collection := db.getAccountsCollection()
	_, err := collection.DeleteOne(db.ctx(), acc)
	if err != nil {
		return err
	}
	return nil
}

func (db *MongoAccountDB) SaveAccount(acc *Account) error {
	collection := db.getAccountsCollection()
	_, err := collection.InsertOne(db.ctx(), acc)
	if err != nil {
		return err
	}
	return nil
}

func (db *MongoAccountDB) FindAccount(acc *Account) (*Account, error) {
	collection := db.getAccountsCollection()
	var result Account

	err := collection.FindOne(db.ctx(), bson.M{"name": acc.Name}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (db *MongoAccountDB) getAccountsCollection() *mongo.Collection {
	return db.Db.Client.Database("Account").Collection("accounts")
}

func (db *MongoAccountDB) ctx() context.Context {
	return context.Background()
}
