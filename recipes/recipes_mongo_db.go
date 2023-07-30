/*
 * MIT License
 *
 * Copyright (c) 2020 Beate Ottenwälder
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

package recipes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strings"
	"sync"

	"github.com/ottenwbe/recipes-manager/utils"
)

const (
	//DATABASE name
	DATABASE = "recipes-manager"
	//RECIPES index
	RECIPES = "recipes"
	//PICTURES index
	PICTURES = "pics"
)

var mongoAddress string

func init() {
	mongoAddress = utils.Config.GetString("recipeDB.host")
}

// MongoRecipeDB implements the Recipe interface to read and write Recipes to and from a Mongo DB
type MongoRecipeDB struct {
	mongoClient *mongo.Client
	//mtx avoids race conditions while connecting to the database and while closing the connection
	mtx sync.Mutex
}

// Clear drops all collections
func (m *MongoRecipeDB) Clear() {
	c := m.getRecipesCollection()
	if err := c.Drop(ctx()); err != nil {
		log.WithError(err).Error("Could not drop recipes from MongoDB")
	}
	r := m.getPictureCollection()
	if err := r.Drop(ctx()); err != nil {
		log.WithError(err).Error("Could not drop pictures from MongoDB")
	}
}

// List all recipes from the db
func (m *MongoRecipeDB) List() (recipes []*Recipe) {

	collection := m.getRecipesCollection()

	recipes = make([]*Recipe, 0)
	cursor, err := collection.Find(ctx(), bson.M{})
	if err != nil {
		log.WithError(err).Info("Error while finding recipe in MongoDB")
		return
	}
	defer func() { _ = cursor.Close(ctx()) }()
	err = cursor.All(ctx(), &recipes)
	if err != nil {
		log.WithError(err).Info("Error while finding recipe in MongoDB")
		return
	}
	return
}

// Num counts the number of recipes in the db
func (m *MongoRecipeDB) Num() int64 {

	collection := m.getRecipesCollection()

	num, err := collection.CountDocuments(ctx(), bson.M{})
	if err != nil {
		log.WithError(err).Info("Error while counting recipes in MongoDB")
	}

	return num
}

// RecipeToBsonM converts a RecipeSearchFilter to a search query (bson.M)
func RecipeToBsonM(searchQuery *RecipeSearchFilter) bson.M {
	query := bson.M{}

	var queryPart = make([]bson.M, 0)

	if searchQuery.Name != "" {
		queryPart = append(queryPart, bson.M{"name": bson.M{"$regex": searchQuery.Name}})
	}
	if searchQuery.Description != "" {
		queryPart = append(queryPart, bson.M{"description": bson.M{"$regex": searchQuery.Description}})
	}
	if searchQuery.Ingredient != nil && len(searchQuery.Ingredient) > 0 {
		rgx := fmt.Sprintf("(%v)", strings.Join(searchQuery.Ingredient, "|"))
		queryPart = append(queryPart, bson.M{"description": bson.M{"$regex": rgx}})
	}

	if len(queryPart) > 1 {
		query["$or"] = queryPart
	} else if len(queryPart) == 1 {
		query = queryPart[0]
	}

	return query
}

// IDs lists all ids of all recipes
func (m *MongoRecipeDB) IDs(searchQuery *RecipeSearchFilter) RecipeList {

	collection := m.getRecipesCollection()

	recipes := make([]*Recipe, 0)
	result := make([]string, 0)

	dbSearch := RecipeToBsonM(searchQuery)

	findOptions := options.Find()
	findOptions.SetProjection(bson.M{"id": 1}) //only get id field

	bsonS, _ := json.Marshal(dbSearch)
	log.WithField("json", string(bsonS)).Debug("Query for IDs")

	cursor, err := collection.Find(ctx(), dbSearch, findOptions)
	if err != nil {
		log.WithError(err).Info("Error while finding recipe")
	}
	defer func() { _ = cursor.Close(ctx()) }()
	err = cursor.All(ctx(), &recipes)

	log.Debugf("Found %v recipes", len(recipes))

	for _, recipe := range recipes {
		result = append(result, recipe.ID.String())
	}

	return RecipeList{Recipes: result}
}

// Get a recipe by ID
func (m *MongoRecipeDB) Get(id RecipeID) *Recipe {

	collection := m.getRecipesCollection()

	recipe := NewInvalidRecipe()
	result := collection.FindOne(ctx(), bson.M{"id": id})

	err := result.Decode(recipe)
	if err != nil {
		log.WithError(err).Info("Error while finding recipe")
	}

	return recipe
}

// Pictures returns all pictures for a given recipe
func (m *MongoRecipeDB) Pictures(id RecipeID) map[string]*RecipePicture {

	collection := m.getPictureCollection()

	recipePictures := make([]*RecipePicture, 0)
	result := make(map[string]*RecipePicture, 0)

	cursor, err := collection.Find(ctx(), bson.M{"id": id})
	if err != nil {
		log.WithError(err).Info("Error while finding recipe pictures")
		return result
	}
	defer func() { _ = cursor.Close(ctx()) }()

	err = cursor.All(ctx(), &recipePictures)
	if err != nil {
		log.WithError(err).Info("Error while finding recipe pictures")
		return result
	}

	for _, recipePicture := range recipePictures {
		result[recipePicture.Name] = recipePicture
	}

	return result
}

// Remove removes a recipe by id
func (m *MongoRecipeDB) Remove(id RecipeID) error {
	c := m.getRecipesCollection()

	_, err := c.DeleteOne(ctx(), bson.M{"id": id})

	return err
}

// Picture returns a specific picture with a specific name for a specific recipe
func (m *MongoRecipeDB) Picture(id RecipeID, name string) *RecipePicture {

	collection := m.getPictureCollection()

	recipePicture := NewInvalidRecipePicture()
	dbResult := collection.FindOne(ctx(), bson.M{"id": id, "name": name})

	err := dbResult.Decode(recipePicture)
	if err != nil {
		log.WithError(err).Error("Error while finding recipe picture")
	}

	return recipePicture
}

// AddPicture to the database
func (m *MongoRecipeDB) AddPicture(pic *RecipePicture) error {

	collection := m.getPictureCollection()

	recipe := m.Get(pic.ID)
	if recipe.ID == InvalidRecipeID() {
		return errors.New("could not find recipe")
	}

	recipe.PictureLink = utils.UniqueSlice(append(recipe.PictureLink, pic.Name))

	err := m.Update(recipe.ID, recipe)
	if err != nil {
		log.WithError(err).Error("Could not insert picture")
		return err
	}

	_, err = collection.InsertOne(ctx(), *pic)
	if err != nil {
		log.WithError(err).Error("Could not insert picture")
		return err
	}

	return nil
}

// Random picture will be returned
func (m *MongoRecipeDB) Random() *Recipe {

	collection := m.getRecipesCollection()

	cursor, err := collection.Aggregate(ctx(), []bson.M{{"$sample": bson.M{"size": 1}}})
	if err != nil {
		log.WithError(err).Info("Error while finding recipe in MongoDB")
		return NewInvalidRecipe()
	}
	defer func() { _ = cursor.Close(ctx()) }()

	recipes := make([]*Recipe, 0)
	err = cursor.All(ctx(), &recipes)
	if err != nil || len(recipes) == 0 {
		log.WithError(err).Info("Error while converting recipe from MongoDB")
		return NewInvalidRecipe()
	}

	return recipes[0]
}

// Update a recipe with a given recipe id
func (m *MongoRecipeDB) Update(id RecipeID, recipe *Recipe) error {

	collection := m.getRecipesCollection()

	_, err := collection.ReplaceOne(ctx(), bson.M{"id": id}, recipe)
	if err != nil {
		log.WithError(err).Error("Could not update recipe")
		return err
	}

	return nil
}

// Insert a recipe into the database
func (m *MongoRecipeDB) Insert(recipe *Recipe) error {

	collection := m.getRecipesCollection()

	_, err := collection.InsertOne(ctx(), *recipe)
	if err != nil {
		log.WithError(err).Error("Could not insert recipe")
		return err
	}

	return nil
}

// Ping MongoDB
func (m *MongoRecipeDB) Ping() error {
	return m.mongoClient.Ping(ctx(), readpref.Primary())
}

// RemoveByName a recipe by name
func (m *MongoRecipeDB) RemoveByName(name string) error {
	c := m.getRecipesCollection()

	_, err := c.DeleteOne(ctx(), bson.M{"name": name})

	return err
}

// GetByName a recipe from the database
func (m *MongoRecipeDB) GetByName(name string) (*Recipe, error) {

	collection := m.getRecipesCollection()

	recipe := *NewInvalidRecipe()
	cur := collection.FindOne(ctx(), bson.M{"name": name})
	err := cur.Decode(&recipe)

	return &recipe, err
}

// Close the connection to the database
func (m *MongoRecipeDB) Close() error {
	return m.StopDB()
}

// StartDB initializes the database connection
func (m *MongoRecipeDB) StartDB() error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.mongoClient == nil {

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
func (m *MongoRecipeDB) StopDB() (err error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.mongoClient != nil {
		err = m.mongoClient.Disconnect(ctx())
	}
	m.mongoClient = nil

	return
}

func (m *MongoRecipeDB) connectToDB() (err error) {
	log.WithField("addr", mongoAddress).Info("Connecting to DB")
	m.mongoClient, err = mongo.NewClient(options.Client().ApplyURI(mongoAddress))
	if err != nil {
		log.WithError(err).Info("Could not create MongoDB client")
		return
	}
	err = m.mongoClient.Connect(ctx())
	if err != nil {
		log.WithError(err).Info("Could not connect to MongoDB")
		return
	}
	err = m.Ping()
	if err != nil {
		log.WithError(err).Info("Could not ping MongoDB")
		return
	}

	err = m.ensureRecipeIndex()
	if err != nil {
		log.WithError(err).Info("Could not create mongo db recipe index")
		return
	}
	err = m.ensurePictureIndex()
	if err != nil {
		log.WithError(err).Info("Could not create mongo db picture index")
		return
	}

	return
}

func (m *MongoRecipeDB) ensureRecipeIndex() error {

	c := m.getRecipesCollection()

	err := m.createTextIndex(c)
	if err != nil {
		return err
	}

	err = m.createDefaultRecipeIndex(c)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoRecipeDB) createDefaultRecipeIndex(c *mongo.Collection) error {
	indexName := mongo.IndexModel{
		Keys: bson.M{ // index in ascending order
			"name": 1,
		},
		Options: options.Index().SetUnique(false).SetSparse(true),
	}

	indexID := mongo.IndexModel{
		Keys: bson.M{ // index in ascending order
			"id": 1,
		},
		Options: options.Index().SetUnique(true).SetSparse(true),
	}

	_, err := c.Indexes().CreateMany(ctx(), []mongo.IndexModel{indexID, indexName})
	if err != nil {
		log.Info("idx error", err)
	}
	return err
}

func (m *MongoRecipeDB) createTextIndex(c *mongo.Collection) error {
	textIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
			{Key: "description", Value: 1},
			{Key: "ingredients.name", Value: 1},
		},
		Options: options.Index().SetUnique(false),
	}

	_, err := c.Indexes().CreateOne(ctx(), textIndex)

	return err
}

func (m *MongoRecipeDB) ensurePictureIndex() error {

	c := m.getPictureCollection()
	index := mongo.IndexModel{
		Keys: bson.M{
			"name": 1, // index in ascending order
		},
		Options: options.Index().SetUnique(false).SetSparse(true),
	}
	_, err := c.Indexes().CreateOne(ctx(), index)

	if err != nil {
		return err
	}

	return nil
}

func (m *MongoRecipeDB) getRecipesCollection() *mongo.Collection {
	return m.mongoClient.Database(DATABASE).Collection(RECIPES)
}

func (m *MongoRecipeDB) getPictureCollection() *mongo.Collection {
	return m.mongoClient.Database(DATABASE).Collection(PICTURES)
}

func ctx() context.Context {
	defaultContext := context.Background()
	return defaultContext
}
