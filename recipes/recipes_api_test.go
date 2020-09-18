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
	"encoding/json"
	"github.com/ottenwbe/go-cook/core"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("recipesAPI", func() {

	var (
		server  core.Server
		recipes RecipeDB
	)

	BeforeSuite(func() {
		handler := core.NewHandler()
		recipes, _ = NewDatabaseClient()
		AddRecipesAPIToHandler(handler, recipes)
		server = core.NewServerA(":8080", handler)
		server.Run()
		time.Sleep(500 * time.Millisecond)
	})

	AfterSuite(func() {
		err := server.Close()
		if err != nil {
			Fail(err.Error())
		}
		err = recipes.Close()
		if err != nil {
			Fail(err.Error())
		}
	})

	Context("Creating the API V1", func() {
		It("should get created", func() {
			resp, err := http.Get("http://localhost:8080/api/v1/recipes")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Context("Get Recipes", func() {
		It("can retrieve an recipe by id", func() {
			expectedRecipe, _ := createRandomRecipes(1, recipes) //recipes

			resp, err := http.Get("http://localhost:8080/api/v1/recipes/r/" + expectedRecipe[0].ID.String())
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))

			var recipe Recipe
			err = json.NewDecoder(resp.Body).Decode(&recipe)
		})

		It("can retrieve an recipe by id and scale the recipe", func() {
			id := NewRecipeID()
			expectedRecipe := NewRecipe(id)
			expectedRecipe.Name = "retrieve recipe"
			expectedRecipe.Ingredients = make([]Ingredients, 0)
			expectedRecipe.Portions = 1
			expectedRecipe.Ingredients = append(expectedRecipe.Ingredients,
				Ingredients{Amount: 100,
					Unit: "g",
					Name: "Test"})
			recipes.Insert(expectedRecipe)

			resp, err := http.Get("http://localhost:8080/api/v1/recipes/r/" + id.String() + "?servings=2")
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))

			var recipe Recipe
			err = json.NewDecoder(resp.Body).Decode(&recipe)
			Expect(len(recipe.Ingredients)).ToNot(Equal(0))
			Expect(recipe.Ingredients[0].Amount).To(Equal(200.0))
		})
	})

	Context("Randomly getting recipes", func() {
		It("returns a 404 when no recipe exists ", func() {
			recipes.Clear()

			resp, err := http.Get("http://localhost:8080/api/v1/recipes/rand")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))
		})

		It("returns a valid random recipe", func() {
			recipes.Clear()
			_, expectedRecipesIDs := createRandomRecipes(10, recipes) //recipes

			resp, err := http.Get("http://localhost:8080/api/v1/recipes/rand")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var recipe Recipe
			err = json.NewDecoder(resp.Body).Decode(&recipe)

			Expect(err).ToNot(HaveOccurred())
			Expect(recipe.ID).To(BeElementOf(expectedRecipesIDs))
		})

	})

	Context("Counting Recipes", func() {
		It("returns 0 when the db is empty", func() {
			recipes.Clear()

			resp, err := http.Get("http://localhost:8080/api/v1/recipes/num")
			Expect(err).ToNot(HaveOccurred())

			result, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))
			Expect(string(result)).To(Equal("0"))
		})
	})

})

func createRandomRecipes(num int, recipes RecipeDB) ([]*Recipe, []RecipeID) {

	randomRecipes := make([]*Recipe, num)
	randomIds := make([]RecipeID, num)

	for id := 0; id < num; id++ {
		randomIds[id] = NewRecipeID()
		randomRecipes[id] = NewRecipe(randomIds[id])
	}

	for _, id := range randomRecipes {
		if err := recipes.Insert(id); err != nil {
			Fail(err.Error())
		}
	}

	return randomRecipes, randomIds
}
