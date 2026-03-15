/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			err error
			db  RecipeDB
		)

		BeforeEach(func() {
			db, err = NewDatabaseClient()
		})

		AfterEach(func() {
			// clean db for testing
			c, cancel := ctx()
			db.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("pics").Drop(c)
			cancel()
			c, cancel = ctx()
			db.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("recipes").Drop(c)
			cancel()
			err = db.Close()
		})

		It("can be established", func() {
			Expect(err).To(BeNil())
		})

		It("can ping the db", func() {
			err = db.Ping()
			Expect(err).To(BeNil())
		})

	})

	Context("picture collection", func() {
		var (
			err         error
			db          RecipeDB
			testRecipe1 *Recipe
			testRecipe2 *Recipe
		)

		prepareTestRecipes := func() {
			testRecipe1 = NewRecipe(NewRecipeID())
			db.Insert(testRecipe1)
			testRecipe2 = NewRecipe(NewRecipeID())
			db.Insert(testRecipe2)
		}

		BeforeEach(func() {
			db, err = NewDatabaseClient()
			prepareTestRecipes()
		})

		AfterEach(func() {
			// clean db for testing
			c, cancel := ctx()
			db.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("pics").Drop(c)
			cancel()
			c, cancel = ctx()
			db.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("recipes").Drop(c)
			cancel()
			db.Close()
		})

		It("can insert a Picture and then read it", func() {
			expectedResult := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = db.AddPicture(expectedResult)
			Expect(err).To(BeNil())
			pic := db.Picture(expectedResult.ID, expectedResult.Name)
			Expect(pic).NotTo(BeNil())
			Expect(pic).To(Equal(expectedResult))
		})

		It("updates the recipe's picturelink", func() {
			expectedResult := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = db.AddPicture(expectedResult)
			Expect(err).To(BeNil())
			recipe := db.Get(testRecipe1.ID)
			Expect(recipe.PictureLink).To(ContainElement(expectedResult.Name))
		})

		It("can insert multiple pictures and then read them", func() {
			add := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = db.AddPicture(add)
			expectedResult := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic2",
				Picture: "thisisabas64picture2",
			}
			err = db.AddPicture(expectedResult)
			Expect(err).To(BeNil())
			pic := db.Picture(expectedResult.ID, expectedResult.Name)
			Expect(pic).To(Equal(expectedResult))
			Expect(pic).NotTo(Equal(add))
		})

		It("can find a picture based on id and name although the name is given multiple times", func() {
			expectedResult1 := &RecipePicture{
				ID:      testRecipe1.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = db.AddPicture(expectedResult1)
			expectedResult2 := &RecipePicture{
				ID:      testRecipe2.ID,
				Name:    "testRecipePic",
				Picture: "thisisabas64picture",
			}
			err = db.AddPicture(expectedResult2)
			Expect(err).To(BeNil())
			pics := db.Pictures(expectedResult1.ID)
			Expect(pics).To(HaveLen(1))
		})

	})

	Context("recipes collection", func() {

		var (
			err error
			db  RecipeDB
		)

		BeforeEach(func() {
			db, err = NewDatabaseClient()
		})

		AfterEach(func() {
			// clean db for testing
			c, cancel := ctx()
			db.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("pics").Drop(c)
			cancel()
			c, cancel = ctx()
			db.(*MongoRecipeDB).MongoClient().Client.Database("recipes-manager").Collection("recipes").Drop(c)
			cancel()

			db.Close()
		})

		It("can insert a Recipe and then read it", func() {
			expectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "testRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the test recipe",
				PictureLink: []string{},
			}
			err = db.Insert(expectedResult)
			Expect(err).To(BeNil())
			recipe, err := db.GetByName(expectedResult.Name)
			defer db.RemoveByName(expectedResult.Name)
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
			err = db.Insert(testInput)
			Expect(err).To(BeNil())
			err := db.Remove(testInput.ID)
			Expect(err).To(BeNil())

			// Try to find it after it has been removed ...
			recipe, err := db.GetByName(testInput.Name)
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
			err = db.Insert(testInput)
			Expect(err).To(BeNil())
			err := db.RemoveByName(testInput.Name)
			Expect(err).To(BeNil())

			// Try to find it after it has been removed ...
			recipe, err := db.GetByName(testInput.Name)
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
			db.Insert(expectedResult)
			defer db.RemoveByName(expectedResult.Name)
			expectedResult2 := &Recipe{
				ID:          NewRecipeID(),
				Name:        "something",
				Ingredients: []Ingredients{},
				Description: "to find",
				PictureLink: []string{},
			}
			db.Insert(expectedResult2)
			defer db.RemoveByName(expectedResult2.Name)

			recipes := db.IDs(&RecipeSearchFilter{Name: "something", Description: "describes"})

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
			db.Insert(expectedResult)
			defer db.RemoveByName(expectedResult.Name)
			unExpectedResult := &Recipe{
				ID:          NewRecipeID(),
				Name:        "noValidTestRecipe",
				Ingredients: []Ingredients{},
				Description: "none",
				PictureLink: []string{},
			}
			db.Insert(unExpectedResult)
			defer db.RemoveByName(unExpectedResult.Name)

			recipes := db.IDs(&RecipeSearchFilter{Description: "describes"})

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
			db.Insert(expectedResult)
			defer db.RemoveByName(expectedResult.Name)

			recipes := db.IDs(&RecipeSearchFilter{Name: "test"})

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
			db.Insert(expectedResult)
			defer db.RemoveByName(expectedResult.Name)

			recipes := db.List()

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
			db.Insert(expectedResult)
			defer db.RemoveByName(expectedResult.Name)

			n := db.Num()

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
			db.Insert(expectedResult)
			defer db.RemoveByName(expectedResult.Name)

			r := db.Random()

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
			db.Insert(expectedResult)
			defer db.RemoveByName(expectedResult.Name)

			names := db.IDs(&RecipeSearchFilter{})

			Expect(names.Recipes).To(ContainElement(expectedResult.ID.String()))
		})

	})
})
