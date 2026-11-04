// SPDX-License-Identifier: MIT

package recipes

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ottenwbe/recipes-manager/config"
	"github.com/ottenwbe/recipes-manager/internal/db"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var _ = Describe("recipes db", func() {

	Context("helper", func() {
		It("can transform Recipe Query to BSON", func() {
			expectedResult := bson.M{}
			result := RecipeToBsonM(&RecipeSearchFilter{})

			Expect(expectedResult).To(Equal(result))
		})

		It("can transform Recipe Query with 1 search parameter to BSON", func() {
			expectedResult := bson.M{"name": bson.M{"$regex": "hi"}}
			result := RecipeToBsonM(&RecipeSearchFilter{Name: "hi"})

			Expect(expectedResult).To(Equal(result))
		})

		It("can transform Recipe Query with 2 search parameter to BSON", func() {
			expectedResult := bson.M{"$or": []bson.M{
				{"name": bson.M{"$regex": "hi"}},
				{"description": bson.M{"$regex": "there"}}}}
			result := RecipeToBsonM(&RecipeSearchFilter{Name: "hi", Description: "there"})

			Expect(expectedResult).To(Equal(result))
		})
	})

	Context("connection", func() {
		var (
			err      error
			coreDB   db.DB
			recipeDB RecipeDB
		)

		BeforeEach(func() {
			addr := config.Config.GetString(db.RecipeDBHostCfg)
			coreDB, err = db.NewDatabaseClient(addr)
			Expect(err).ToNot(HaveOccurred())
			recipeDB, err = NewRecipeDB(coreDB)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			// clean db for testing
			c, cancel := createContext()
			recipeDB.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("pics").Drop(c)
			cancel()
			c, cancel = createContext()
			recipeDB.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("recipes").Drop(c)
			cancel()
			err = coreDB.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("can be established", func() {
			Expect(err).To(BeNil())
		})

		It("can ping the db", func() {
			err = recipeDB.Ping()
			Expect(err).To(BeNil())
		})

	})

	Context("picture collection", func() {
		var (
			err         error
			coreDB      db.DB
			recipeDB    RecipeDB
			testRecipe1 *Recipe
			testRecipe2 *Recipe
		)

		prepareTestRecipes := func() {
			testRecipe1 = NewRecipe(NewRecipeID())
			err := recipeDB.Insert(testRecipe1)
			Expect(err).ToNot(HaveOccurred())
			testRecipe2 = NewRecipe(NewRecipeID())
			err = recipeDB.Insert(testRecipe2)
			Expect(err).ToNot(HaveOccurred())
		}

		BeforeEach(func() {
			addr := config.Config.GetString(db.RecipeDBHostCfg)
			coreDB, err = db.NewDatabaseClient(addr)
			Expect(err).ToNot(HaveOccurred())
			recipeDB, err = NewRecipeDB(coreDB)
			Expect(err).ToNot(HaveOccurred())
			prepareTestRecipes()
		})

		AfterEach(func() {
			// clean db for testing
			c, cancel := createContext()
			recipeDB.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("pics").Drop(c)
			cancel()
			c, cancel = createContext()
			recipeDB.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("recipes").Drop(c)
			cancel()
			err = coreDB.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("can insert a Picture and then read it", func() {
			expectedResult := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = recipeDB.AddPicture(expectedResult)
			Expect(err).To(BeNil())
			pic := recipeDB.Picture(expectedResult.ID, expectedResult.Name)
			Expect(pic).NotTo(BeNil())
			Expect(pic).To(Equal(expectedResult))
		})

		It("updates the recipe's picturelink", func() {
			expectedResult := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = recipeDB.AddPicture(expectedResult)
			Expect(err).To(BeNil())
			recipe := recipeDB.Get(testRecipe1.ID)
			Expect(recipe.PictureLink).To(ContainElement(expectedResult.Name))
		})

		It("can insert multiple pictures and then read them", func() {
			add := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = recipeDB.AddPicture(add)
			expectedResult := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic2",
				Picture: "thisisabas64picture2",
			}
			err = recipeDB.AddPicture(expectedResult)
			Expect(err).To(BeNil())
			pic := recipeDB.Picture(expectedResult.ID, expectedResult.Name)
			Expect(pic).To(Equal(expectedResult))
			Expect(pic).NotTo(Equal(add))
		})

		It("can find a picture based on id and name although the name is given multiple times", func() {
			expectedResult1 := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = recipeDB.AddPicture(expectedResult1)
			expectedResult2 := &RecipePicture{
				ID:      testRecipe2.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = recipeDB.AddPicture(expectedResult2)
			Expect(err).To(BeNil())
			pics := recipeDB.Pictures(expectedResult1.ID)
			Expect(pics).To(HaveLen(1))
		})

	})

	Context("recipes collection", func() {

		var (
			err      error
			coreDB   db.DB
			recipeDB RecipeDB
		)

		BeforeEach(func() {
			addr := config.Config.GetString(db.RecipeDBHostCfg)
			coreDB, err = db.NewDatabaseClient(addr)
			Expect(err).ToNot(HaveOccurred())
			recipeDB, err = NewRecipeDB(coreDB)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			// clean db for testing
			c, cancel := createContext()
			recipeDB.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("pics").Drop(c)
			cancel()
			c, cancel = createContext()
			recipeDB.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("recipes").Drop(c)
			cancel()

			err = coreDB.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("can insert a Recipe and then read it", func() {
			expectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "testRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the test recipe",
				PictureLink: []string{},
			}
			err = recipeDB.Insert(expectedResult)
			Expect(err).To(BeNil())
			recipe, err := recipeDB.GetByName(expectedResult.Name)
			defer recipeDB.RemoveByName(expectedResult.Name)
			Expect(err).To(BeNil())
			Expect(recipe).To(Equal(expectedResult))
		})

		It("can remove a Recipe by id", func() {
			testInput := &Recipe{
				ID:          NewRecipeID(),
				Name:        "removeIDTestRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the remove test recipe",
				PictureLink: []string{},
			}
			err = recipeDB.Insert(testInput)
			Expect(err).To(BeNil())
			err := recipeDB.Remove(testInput.ID)
			Expect(err).To(BeNil())

			// Try to find it after it has been removed ...
			recipe, err := recipeDB.GetByName(testInput.Name)
			Expect(err).ToNot(BeNil())
			Expect(recipe).ToNot(Equal(testInput))
		})

		It("can remove a Recipe by name", func() {
			testInput := &Recipe{
				ID:          NewRecipeID(),
				Name:        "removeTestRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the remove test recipe",
				PictureLink: []string{},
			}
			err = recipeDB.Insert(testInput)
			Expect(err).To(BeNil())
			err := recipeDB.RemoveByName(testInput.Name)
			Expect(err).To(BeNil())

			// Try to find it after it has been removed ...
			recipe, err := recipeDB.GetByName(testInput.Name)
			Expect(err).ToNot(BeNil())
			Expect(recipe).ToNot(Equal(testInput))
		})

		It("can list all Recipes and filter them by name and description", func() {
			expectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "testRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the test recipe",
				PictureLink: []string{},
			}
			recipeDB.Insert(expectedResult)
			defer recipeDB.RemoveByName(expectedResult.Name)
			expectedResult2 := &Recipe{
				ID:          NewRecipeID(),
				Name:        "something",
				Ingredients: []Ingredients{},
				Description: "to find",
				PictureLink: []string{},
			}
			recipeDB.Insert(expectedResult2)
			defer recipeDB.RemoveByName(expectedResult2.Name)

			recipes := recipeDB.IDs(&RecipeSearchFilter{Name: "something", Description: "describes"})

			Expect(err).To(BeNil())
			Expect(recipes.Recipes).To(ContainElement(expectedResult.ID.String()))
			Expect(recipes.Recipes).To(ContainElement(expectedResult2.ID.String()))
		})

		It("can list all Recipes and filter them by description", func() {
			expectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "testRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the test recipe",
				PictureLink: []string{},
			}
			recipeDB.Insert(expectedResult)
			defer recipeDB.RemoveByName(expectedResult.Name)
			unExpectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "noValidTestRecipe",
				Ingredients: []Ingredients{},
				Description: "none",
				PictureLink: []string{},
			}
			recipeDB.Insert(unExpectedResult)
			defer recipeDB.RemoveByName(unExpectedResult.Name)

			recipes := recipeDB.IDs(&RecipeSearchFilter{Description: "describes"})

			Expect(err).To(BeNil())
			Expect(recipes.Recipes).To(ContainElement(expectedResult.ID.String()))
		})

		It("can list all Recipes and filter them by name", func() {
			expectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "testRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the test recipe",
				PictureLink: []string{},
			}
			recipeDB.Insert(expectedResult)
			defer recipeDB.RemoveByName(expectedResult.Name)

			recipes := recipeDB.IDs(&RecipeSearchFilter{Name: "test"})

			Expect(err).To(BeNil())
			Expect(recipes.Recipes).To(ContainElement(expectedResult.ID.String()))
		})

		It("can list all Recipes", func() {
			expectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "testRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the test recipe",
				PictureLink: []string{},
			}
			recipeDB.Insert(expectedResult)
			defer recipeDB.RemoveByName(expectedResult.Name)

			recipes := recipeDB.List()

			Expect(err).To(BeNil())
			Expect(recipes).To(ContainElement(expectedResult))
		})

		It("can count up all elements", func() {
			expectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "testRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the test recipe",
				PictureLink: []string{},
			}
			recipeDB.Insert(expectedResult)
			defer recipeDB.RemoveByName(expectedResult.Name)

			n := recipeDB.Num()

			Expect(n).To(BeNumerically("==", 1))
		})

		It("can get a Recipe at random", func() {
			expectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "testRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the test recipe",
				PictureLink: []string{},
			}
			recipeDB.Insert(expectedResult)
			defer recipeDB.RemoveByName(expectedResult.Name)

			r := recipeDB.Random()

			Expect(err).To(BeNil())
			Expect(r).To(Equal(expectedResult))
		})

		It("can aggregate the names of all elements", func() {
			expectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "testRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the test recipe",
				PictureLink: []string{},
			}
			recipeDB.Insert(expectedResult)
			defer recipeDB.RemoveByName(expectedResult.Name)

			names := recipeDB.IDs(&RecipeSearchFilter{})

			Expect(names.Recipes).To(ContainElement(expectedResult.ID.String()))
		})

	})
})
