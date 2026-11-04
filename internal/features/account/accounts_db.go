// SPDX-License-Identifier: MIT

package account

import (
	"context"
	"time"

	"github.com/ottenwbe/recipes-manager/config"
	"github.com/ottenwbe/recipes-manager/internal/db"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoAccountService stores and manipulates stored accounts in a DB
type MongoAccountService struct {
	DbClient *db.MongoClient
}

const (
	// EMAIL key
	EMAIL = "name"
	// ID key
	ID = "id"
	// AccountDBNameCfg is the config key for the account database name
	AccountDBNameCfg = "accountDB.name"
	// AccountDBCollectionCfg is the config key for the account collection name
	AccountDBCollectionCfg = "accountDB.collection"
)

func init() {
	config.Config.SetDefault(AccountDBNameCfg, "accounts")
	config.Config.SetDefault(AccountDBCollectionCfg, "accounts")
}

// NewMongoAccountService is created and configured
func NewMongoAccountService(database db.DB) *MongoAccountService {

	accountDB := &MongoAccountService{
		DbClient: database.(*db.MongoClient),
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
	ctx, cancel := db.ctx()
	defer cancel()
	_, err := collection.DeleteOne(ctx, bson.M{ID: id})
	if err != nil {
		return err
	}
	return nil
}

// DeleteAccountByName deletes an account indefinitely
func (db *MongoAccountService) DeleteAccountByName(name string) error {
	collection := db.getAccountsCollection()
	ctx, cancel := db.ctx()
	defer cancel()
	_, err := collection.DeleteOne(ctx, bson.M{EMAIL: name})
	if err != nil {
		return err
	}
	return nil
}

// NewAccount stores the account information in a db
func (db *MongoAccountService) NewAccount(name string, t Type) (*Account, error) {
	collection := db.getAccountsCollection()

	acc := NewAccount(name, t)

	ctx, cancel := db.ctx()
	defer cancel()
	_, err := collection.InsertOne(ctx, acc)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

// FindAccount by Name
func (db *MongoAccountService) FindAccount(name string) (*Account, error) {
	collection := db.getAccountsCollection()
	var result Account

	ctx, cancel := db.ctx()
	defer cancel()
	err := collection.FindOne(ctx, bson.M{EMAIL: name}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (db *MongoAccountService) getAccountsCollection() *mongo.Collection {
	dbName := config.Config.GetString(AccountDBNameCfg)
	collName := config.Config.GetString(AccountDBCollectionCfg)
	return db.DbClient.Client.Database(dbName).Collection(collName)
}

func (*MongoAccountService) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func (db *MongoAccountService) createTextIndex() error {

	c := db.getAccountsCollection()

	textIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: EMAIL, Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	ctx, cancel := db.ctx()
	defer cancel()
	_, err := c.Indexes().CreateOne(ctx, textIndex)

	return err
}
