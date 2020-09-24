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
	"bytes"
	"encoding/json"
	"fmt"
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
			id := createAndPersistDefaultRecipe(recipes)

			resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/recipes/r/%v?servings=2", id.String()))
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

		It("can retrieve a random recipe and scale the recipe", func() {
			recipes.Clear()

			_ = createAndPersistDefaultRecipe(recipes)

			resp, err := http.Get("http://localhost:8080/api/v1/recipes/rand?servings=2")
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))

			var recipe Recipe
			err = json.NewDecoder(resp.Body).Decode(&recipe)
			Expect(len(recipe.Ingredients)).ToNot(Equal(0))
			Expect(recipe.Ingredients[0].Amount).To(Equal(200.0))
		})
	})

	Context("Counting Recipes", func() {
		It("returns 0 when no recipes are persisted", func() {
			recipes.Clear()

			_, _ = createRandomRecipes(10, recipes)

			resp, err := http.Get("http://localhost:8080/api/v1/recipes/num")
			Expect(err).ToNot(HaveOccurred())

			result, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))
			Expect(string(result)).To(Equal("10"))
		})

		It("returns the amount of persisted recipes", func() {
			recipes.Clear()

			resp, err := http.Get("http://localhost:8080/api/v1/recipes/num")
			Expect(err).ToNot(HaveOccurred())

			result, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))
			Expect(string(result)).To(Equal("0"))
		})
	})

	Context("Posting Recipes", func() {
		It("is not possible with malformed documents", func() {
			recipes.Clear()

			resp, err := http.Post("http://localhost:8080/api/v1/recipes", "application/json", bytes.NewBuffer(nil))
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(400))
		})

		It("persists a new recipe", func() {
			recipes.Clear()

			const POSTTEST = "PostTest"
			const POSTDESCRIPTION = "Test \n 123"

			recipe := Recipe{Servings: 2, Name: POSTTEST, Description: POSTDESCRIPTION}
			recipeJSON, _ := json.Marshal(recipe)

			resp, err := http.Post("http://localhost:8080/api/v1/recipes", "application/json", bytes.NewBuffer(recipeJSON))
			Expect(err).ToNot(HaveOccurred())

			retrievedRecipe, err := recipes.GetByName(POSTTEST)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))
			Expect(retrievedRecipe.Servings).To(Equal(recipe.Servings))
			Expect(retrievedRecipe.Description).To(Equal(recipe.Description))
		})
	})

	Context("DELETE Recipes", func() {

		It("removes a persisted recipe", func() {
			id := createAndPersistDefaultRecipe(recipes)
			client := &http.Client{}
			request, err := http.NewRequest(http.MethodDelete, "http://localhost:8080/api/v1/recipes/r/"+id.String(), bytes.NewBuffer(nil))
			response, err := client.Do(request)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("PUT Recipes", func() {

		It("persists a change to a recipe", func() {
			recipes.Clear()

			const POSTTEST = "PostTest"
			const POSTDESCRIPTION = "Test \n 123"

			recipe := Recipe{Servings: 2, Name: POSTTEST, Description: POSTDESCRIPTION}
			recipeJSON, _ := json.Marshal(recipe)

			_, err := http.Post("http://localhost:8080/api/v1/recipes", "application/json", bytes.NewBuffer(recipeJSON))
			Expect(err).ToNot(HaveOccurred())

			// update recipe
			recipe.Servings = 3
			recipe.Description = "updated"

			// post updated recipe
			client := &http.Client{}
			recipeJSON, _ = json.Marshal(recipe)
			retrievedRecipe, _ := recipes.GetByName(POSTTEST)
			request, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/recipes/r/"+retrievedRecipe.ID.String(), bytes.NewBuffer(recipeJSON))
			request.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(request)
			Expect(err).ToNot(HaveOccurred())

			retrievedRecipe, err = recipes.GetByName(POSTTEST)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(204))
			Expect(retrievedRecipe.Servings).To(Equal(recipe.Servings))
			Expect(retrievedRecipe.Description).To(Equal(recipe.Description))
		})
	})

})

func createAndPersistDefaultRecipe(recipes RecipeDB) RecipeID {
	id := NewRecipeID()

	expectedRecipe := NewRecipe(id)
	expectedRecipe.Name = "retrieve recipe"
	expectedRecipe.Ingredients = make([]Ingredients, 0)
	expectedRecipe.Servings = 1
	expectedRecipe.Ingredients = append(expectedRecipe.Ingredients,
		Ingredients{Amount: 100,
			Unit: "g",
			Name: "Test"})
	recipes.Insert(expectedRecipe)

	return id
}

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
