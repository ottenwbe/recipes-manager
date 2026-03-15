/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ottenwbe/recipes-manager/core"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	// DATABASE name
	DATABASE = "recipes-manager"
	// RECIPES index
	RECIPES = "recipes"
	// PICTURES index
	PICTURES = "pics"
)

// MongoRecipeDB implements the Recipe interface to read and write Recipes to and from a Mongo DB
type MongoRecipeDB struct {
	mongoClient *core.MongoClient
}

// MongoClient returns the underlying core.MongoClient
func (m *MongoRecipeDB) MongoClient() *core.MongoClient {
	return m.mongoClient
}

// Clear drops all collections
func (m *MongoRecipeDB) Clear() {
	c := m.getRecipesCollection()
	co, cancel := ctx()
	if err := c.Drop(co); err != nil {
		log.WithError(err).Error("Could not drop recipes from MongoDB")
	}
	cancel()
	r := m.getPictureCollection()
	co, cancel = ctx()
	defer cancel()
	if err := r.Drop(co); err != nil {
		log.WithError(err).Error("Could not drop pictures from MongoDB")
	}
}

// List all recipes from the db
func (m *MongoRecipeDB) List() (recipes []*Recipe) {

	collection := m.getRecipesCollection()

	recipes = make([]*Recipe, 0)
	c, cancel := ctx()
	defer cancel()
	cursor, err := collection.Find(c, bson.M{})
	if err != nil {
		log.WithError(err).Info("Error while finding recipe in MongoDB")
		return
	}
	defer func() { _ = cursor.Close(c) }()
	err = cursor.All(c, &recipes)
	if err != nil {
		log.WithError(err).Info("Error while finding recipe in MongoDB")
		return
	}
	return
}

// Num counts the number of recipes in the db
func (m *MongoRecipeDB) Num() int64 {

	collection := m.getRecipesCollection()

	c, cancel := ctx()
	defer cancel()
	num, err := collection.CountDocuments(c, bson.M{})
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
	if len(searchQuery.Ingredient) > 0 {
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
	findOptions.SetProjection(bson.M{"id": 1}) // only get id field

	bsonS, _ := json.Marshal(dbSearch)
	log.WithField("json", string(bsonS)).Debug("Query for IDs")

	c, cancel := ctx()
	defer cancel()
	cursor, err := collection.Find(c, dbSearch, findOptions)
	if err != nil {
		log.WithError(err).Info("Error while finding recipe")
	}
	defer func() { _ = cursor.Close(c) }()
	err = cursor.All(c, &recipes)
	if err != nil {
		log.WithError(err).Info("Error while finding recipe")
	}

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
	c, cancel := ctx()
	defer cancel()
	result := collection.FindOne(c, bson.M{"id": id})

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

	c, cancel := ctx()
	defer cancel()
	cursor, err := collection.Find(c, bson.M{"id": id})
	if err != nil {
		log.WithError(err).Info("Error while finding recipe pictures")
		return result
	}
	defer func() { _ = cursor.Close(c) }()

	err = cursor.All(c, &recipePictures)
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

	ctx, cancel := ctx()
	defer cancel()
	_, err := c.DeleteOne(ctx, bson.M{"id": id})

	return err
}

// Picture returns a specific picture with a specific name for a specific recipe
func (m *MongoRecipeDB) Picture(id RecipeID, name string) *RecipePicture {

	collection := m.getPictureCollection()

	recipePicture := NewInvalidRecipePicture()
	c, cancel := ctx()
	defer cancel()
	dbResult := collection.FindOne(c, bson.M{"id": id, "name": name})

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

	recipe.PictureLink = UniqueSlice(append(recipe.PictureLink, pic.Name))

	err := m.Update(recipe.ID, recipe)
	if err != nil {
		log.WithError(err).Error("Could not insert picture")
		return err
	}

	c, cancel := ctx()
	defer cancel()
	_, err = collection.InsertOne(c, *pic)
	if err != nil {
		log.WithError(err).Error("Could not insert picture")
		return err
	}

	return nil
}

// Random picture will be returned
func (m *MongoRecipeDB) Random() *Recipe {

	collection := m.getRecipesCollection()

	c, cancel := ctx()
	defer cancel()
	cursor, err := collection.Aggregate(c, []bson.M{{"$sample": bson.M{"size": 1}}})
	if err != nil {
		log.WithError(err).Info("Error while finding recipe in MongoDB")
		return NewInvalidRecipe()
	}
	defer func() { _ = cursor.Close(c) }()

	recipes := make([]*Recipe, 0)
	err = cursor.All(c, &recipes)
	if err != nil || len(recipes) == 0 {
		log.WithError(err).Info("Error while converting recipe from MongoDB")
		return NewInvalidRecipe()
	}

	return recipes[0]
}

// Update a recipe with a given recipe id
func (m *MongoRecipeDB) Update(id RecipeID, recipe *Recipe) error {

	collection := m.getRecipesCollection()

	c, cancel := ctx()
	defer cancel()
	_, err := collection.ReplaceOne(c, bson.M{"id": id}, recipe)
	if err != nil {
		log.WithError(err).Error("Could not update recipe")
		return err
	}

	return nil
}

// Insert a recipe into the database
func (m *MongoRecipeDB) Insert(recipe *Recipe) error {

	collection := m.getRecipesCollection()

	c, cancel := ctx()
	defer cancel()
	_, err := collection.InsertOne(c, *recipe)
	if err != nil {
		log.WithError(err).Error("Could not insert recipe")
		return err
	}

	return nil
}

// Ping MongoDB
func (m *MongoRecipeDB) Ping() error {
	return m.mongoClient.Ping()
}

// RemoveByName a recipe by name
func (m *MongoRecipeDB) RemoveByName(name string) error {
	c := m.getRecipesCollection()

	ctx, cancel := ctx()
	defer cancel()
	_, err := c.DeleteOne(ctx, bson.M{"name": name})

	return err
}

// GetByName a recipe from the database
func (m *MongoRecipeDB) GetByName(name string) (*Recipe, error) {

	collection := m.getRecipesCollection()

	recipe := *NewInvalidRecipe()
	c, cancel := ctx()
	defer cancel()
	cur := collection.FindOne(c, bson.M{"name": name})
	err := cur.Decode(&recipe)

	return &recipe, err
}

// Close the connection to the database
func (m *MongoRecipeDB) Close() error {
	return m.StopDB()
}

// StartDB initializes the database connection
func (m *MongoRecipeDB) StartDB(addr string) error {
	if m.mongoClient == nil {
		m.mongoClient = &core.MongoClient{}
	}

	err := m.mongoClient.StartDB(addr)
	if err != nil {
		return err
	}

	err = m.ensureRecipeIndex()
	if err != nil {
		log.WithError(err).Info("Could not create mongo db recipe index")
		return err
	}
	err = m.ensurePictureIndex()
	if err != nil {
		log.WithError(err).Info("Could not create mongo db picture index")
		return err
	}

	return nil
}

// StopDB closes the connection to the db
func (m *MongoRecipeDB) StopDB() (err error) {
	return m.mongoClient.StopDB()
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

func (*MongoRecipeDB) createDefaultRecipeIndex(c *mongo.Collection) error {
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

	ctx, cancel := ctx()
	defer cancel()
	_, err := c.Indexes().CreateMany(ctx, []mongo.IndexModel{indexID, indexName})
	if err != nil {
		log.Info("idx error", err)
	}
	return err
}

func (*MongoRecipeDB) createTextIndex(c *mongo.Collection) error {
	textIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
			{Key: "description", Value: 1},
			{Key: "ingredients.name", Value: 1},
		},
		Options: options.Index().SetUnique(false),
	}

	ctx, cancel := ctx()
	defer cancel()
	_, err := c.Indexes().CreateOne(ctx, textIndex)

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
	ctx, cancel := ctx()
	defer cancel()
	_, err := c.Indexes().CreateOne(ctx, index)

	if err != nil {
		return err
	}

	return nil
}

func (m *MongoRecipeDB) getRecipesCollection() *mongo.Collection {
	return m.mongoClient.Client.Database(DATABASE).Collection(RECIPES)
}

func (m *MongoRecipeDB) getPictureCollection() *mongo.Collection {
	return m.mongoClient.Client.Database(DATABASE).Collection(PICTURES)
}

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
