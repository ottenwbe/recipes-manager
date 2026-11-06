// SPDX-License-Identifier: MIT

package recipes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ottenwbe/recipes-manager/internal/config"
	"github.com/ottenwbe/recipes-manager/internal/db"

	"github.com/ottenwbe/recipes-manager/internal/server"
	"github.com/ottenwbe/recipes-manager/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	appServer *server.Server
	database  db.DB
	recipes   RecipeDB
)

var _ = BeforeSuite(func() {
	handler := server.NewHandler("*")
	var err error
	addr := config.Config.GetString(db.RecipeDBHostCfg)
	database, err = db.NewDatabaseClient(addr)
	Expect(err).ToNot(HaveOccurred())
	recipes, err = NewRecipeDB(database)
	Expect(err).ToNot(HaveOccurred())
	err = AddRecipesAPIToHandler(handler, recipes)
	Expect(err).ToNot(HaveOccurred())
	appServer = server.NewServerWithAddress(":8080", handler)
	appServer.Run()
	time.Sleep(500 * time.Millisecond)
})

var _ = AfterSuite(func() {
	err := appServer.Close()
	if err != nil {
		Fail(err.Error())
	}
	err = database.Close()
	if err != nil {
		Fail(err.Error())
	}
})

var _ = Describe("recipesAPI", func() {

	Context("Creating the API V1", func() {
		It("should get created", func() {
			resp, err := http.Get("http://localhost:8080/api/v1/recipes")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Context("List Recipes", func() {
		It("should be able to filter by name", func() {

			createRandomRecipes(5, recipes)
			expectedID := createAndPersistDefaultRecipe(recipes)
			defer func() { // delete recently created recipe
				recipes.Remove(expectedID)
			}()

			resp, err := http.Get("http://localhost:8080/api/v1/recipes?name=retrieve")

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var recipeIDs models.RecipeList
			_ = json.NewDecoder(resp.Body).Decode(&recipeIDs)
			Expect(len(recipeIDs.Recipes)).To(Equal(1))
			Expect(expectedID.String()).To(Equal(recipeIDs.Recipes[0]))
		})

		It("should be able to filter by description", func() {

			createRandomRecipes(5, recipes)
			expectedID := createAndPersistDefaultRecipe(recipes)
			defer func() { // delete recently created recipe
				recipes.Remove(expectedID)
			}()

			resp, err := http.Get("http://localhost:8080/api/v1/recipes?description=details")

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var recipeIDs models.RecipeList
			_ = json.NewDecoder(resp.Body).Decode(&recipeIDs)
			Expect(len(recipeIDs.Recipes)).To(Equal(1))
			Expect(expectedID.String()).To(Equal(recipeIDs.Recipes[0]))
		})

		It("should be able to filter by name and description", func() {

			createRandomRecipes(5, recipes) //add noise
			expectedID1 := createAndPersistNewRecipe("search", "none", models.Ingredients{Name: "hi"}, recipes)
			defer func() { // delete recently created recipe
				recipes.Remove(expectedID1)
			}()
			expectedID2 := createAndPersistNewRecipe("none", "find", models.Ingredients{Name: "hi"}, recipes)
			defer func() { // delete recently created recipe
				recipes.Remove(expectedID2)
			}()

			resp, err := http.Get("http://localhost:8080/api/v1/recipes?description=find&name=search")

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var recipeIDs models.RecipeList
			_ = json.NewDecoder(resp.Body).Decode(&recipeIDs)
			Expect(len(recipeIDs.Recipes)).To(Equal(2))
			Expect(recipeIDs.Recipes).To(ContainElement(expectedID1.String()))
			Expect(recipeIDs.Recipes).To(ContainElement(expectedID2.String()))
		})

		It("should not return elements that do not match the search query", func() {

			createRandomRecipes(5, recipes) //add noise
			expectedID := createAndPersistNewRecipe("search", "test", models.Ingredients{Name: "hi"}, recipes)
			defer func() { // delete recently created recipe
				recipes.Remove(expectedID)
			}()

			resp, err := http.Get("http://localhost:8080/api/v1/recipes?description=find")

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var recipeIDs []string
			_ = json.NewDecoder(resp.Body).Decode(&recipeIDs)
			Expect(len(recipeIDs)).To(Equal(0))
		})
	})

	Context("Get Recipes", func() {
		It("can retrieve an recipe by id", func() {
			expectedRecipe, _ := createRandomRecipes(1, recipes) //recipes

			resp, err := http.Get("http://localhost:8080/api/v1/recipes/r/" + expectedRecipe[0].ID.String())
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))

			var recipe models.Recipe
			_ = json.NewDecoder(resp.Body).Decode(&recipe)
			Expect(recipe).To(Equal(*expectedRecipe[0]))
		})

		It("can retrieve an recipe by id and scale the recipe", func() {
			id := createAndPersistDefaultRecipe(recipes)

			resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/recipes/r/%v?servings=2", id.String()))
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))

			var recipe models.Recipe
			_ = json.NewDecoder(resp.Body).Decode(&recipe)
			Expect(len(recipe.Ingredients)).ToNot(Equal(0))
			Expect(recipe.Ingredients[0].Amount).To(Equal(200.0))
		})
	})

	Context("Randomly getting recipes", func() {
		It("returns a 404 when no recipe exists ", func() {
			recipes.Clear()

			resp, err := http.Get("http://localhost:8080/api/v1/recipes-random")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))
		})

		It("returns a valid random recipe", func() {
			recipes.Clear()
			_, expectedRecipesIDs := createRandomRecipes(10, recipes) //recipes

			resp, err := http.Get("http://localhost:8080/api/v1/recipes-random")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var recipe models.Recipe
			_ = json.NewDecoder(resp.Body).Decode(&recipe)

			Expect(err).ToNot(HaveOccurred())
			Expect(recipe.ID).To(BeElementOf(expectedRecipesIDs))
		})

		It("can retrieve a random recipe and scale the recipe", func() {
			recipes.Clear()

			_ = createAndPersistDefaultRecipe(recipes)

			resp, err := http.Get("http://localhost:8080/api/v1/recipes-random?servings=2")
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))

			var recipe models.Recipe
			_ = json.NewDecoder(resp.Body).Decode(&recipe)
			Expect(len(recipe.Ingredients)).ToNot(Equal(0))
			Expect(recipe.Ingredients[0].Amount).To(Equal(200.0))
		})
	})

	Context("Counting Recipes", func() {
		It("returns 0 when no recipes are persisted", func() {
			recipes.Clear()

			_, _ = createRandomRecipes(10, recipes)

			resp, err := http.Get("http://localhost:8080/api/v1/recipes-num")
			Expect(err).ToNot(HaveOccurred())

			result, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))
			Expect(string(result)).To(Equal("10"))
		})

		It("returns the amount of persisted recipes", func() {
			recipes.Clear()

			resp, err := http.Get("http://localhost:8080/api/v1/recipes-num")
			Expect(err).ToNot(HaveOccurred())

			result, err := io.ReadAll(resp.Body)
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

			recipe := models.Recipe{Servings: 2, Name: POSTTEST, Description: POSTDESCRIPTION}
			recipeJSON, _ := json.Marshal(recipe)

			resp, err := http.Post("http://localhost:8080/api/v1/recipes", "application/json", bytes.NewBuffer(recipeJSON))
			Expect(err).ToNot(HaveOccurred())

			retrievedRecipe, err := recipes.GetByName(POSTTEST)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(201))
			Expect(retrievedRecipe.Servings).To(Equal(recipe.Servings))
			Expect(retrievedRecipe.Description).To(Equal(recipe.Description))
		})
	})

	Context("DELETE Recipes", func() {

		It("removes a persisted recipe", func() {
			id := createAndPersistDefaultRecipe(recipes)
			client := &http.Client{}
			request, _ := http.NewRequest(http.MethodDelete, "http://localhost:8080/api/v1/recipes/r/"+id.String(), bytes.NewBuffer(nil))
			response, err := client.Do(request)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusNoContent))
		})
	})

	Context("PUT Recipes", func() {

		It("persists a change to a recipe", func() {
			recipes.Clear()

			const POSTTEST = "PostTest"
			const POSTDESCRIPTION = "Test \n 123"

			recipe := models.Recipe{Servings: 2, Name: POSTTEST, Description: POSTDESCRIPTION}
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
			request, _ := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/recipes/r/"+retrievedRecipe.ID.String(), bytes.NewBuffer(recipeJSON))
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

func createAndPersistNewRecipe(name string, description string, ingredient models.Ingredients, recipes RecipeDB) models.RecipeID {
	id := models.NewRecipeID()

	expectedRecipe := models.NewRecipe(id)
	expectedRecipe.Name = name
	expectedRecipe.Ingredients = make([]models.Ingredients, 0)
	expectedRecipe.Servings = 1
	expectedRecipe.Description = description
	expectedRecipe.Ingredients = append(expectedRecipe.Ingredients, ingredient)
	recipes.Insert(expectedRecipe)

	return id
}

func createAndPersistDefaultRecipe(recipes RecipeDB) models.RecipeID {
	ingredient := models.Ingredients{Amount: 100,
		Unit: "g",
		Name: "Test"}
	return createAndPersistNewRecipe("retrieve recipe", "details", ingredient, recipes)
}

func createRandomRecipes(num int, recipes RecipeDB) ([]*models.Recipe, []models.RecipeID) {

	randomRecipes := make([]*models.Recipe, num)
	randomIds := make([]models.RecipeID, num)

	for id := 0; id < num; id++ {
		randomIds[id] = models.NewRecipeID()
		randomRecipes[id] = models.NewRecipe(randomIds[id])
	}

	for _, id := range randomRecipes {
		if err := recipes.Insert(id); err != nil {
			Fail(err.Error())
		}
	}

	return randomRecipes, randomIds
}
