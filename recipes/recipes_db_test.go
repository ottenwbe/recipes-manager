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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("recipes db", func() {

	Context("connection", func() {
		var (
			err error
			db  RecipeDB
		)

		BeforeEach(func() {
			db, err = NewDatabaseClient()
		})

		AfterEach(func() {
			db.Close()
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
			// clean db for testing
			db.(*MongoRecipeDB).mongoClient.Database("go-cook").Collection("pics").Drop(ctx())
			db.(*MongoRecipeDB).mongoClient.Database("go-cook").Collection("recipes").Drop(ctx())
			prepareTestRecipes()
		})

		AfterEach(func() {
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
			// clean db for testing
			db.(*MongoRecipeDB).mongoClient.Database("go-cook").Collection("recipes").Drop(ctx())
		})

		AfterEach(func() {
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
			defer db.Remove(expectedResult.Name)
			Expect(err).To(BeNil())
			Expect(recipe).To(Equal(expectedResult))
		})

		It("can remove a Recipe", func() {
			testInput := &Recipe{
				ID:          NewRecipeID(),
				Name:        "removeTestRecipe",
				Ingredients: []Ingredients{},
				Description: "describes the remove test recipe",
				PictureLink: []string{},
			}
			err = db.Insert(testInput)
			Expect(err).To(BeNil())
			err := db.Remove(testInput.Name)

			// Try to find it after it has been removed ...
			recipe, err := db.GetByName(testInput.Name)
			Expect(err).ToNot(BeNil())
			Expect(recipe).ToNot(Equal(testInput))
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
			defer db.Remove(expectedResult.Name)

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
			defer db.Remove(expectedResult.Name)

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
			defer db.Remove(expectedResult.Name)

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
			defer db.Remove(expectedResult.Name)

			names := db.IDs()

			Expect(names).To(ContainElement(expectedResult.ID.String()))
		})

	})
})
