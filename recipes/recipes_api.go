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
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ottenwbe/go-cook/core"
	log "github.com/sirupsen/logrus"
)

const (
	// SERVINGS keyword used as part of the url
	SERVINGS = "servings"
	// RECIPE keyword used as part of the url
	RECIPE = "recipe"
	// NAME keyword used as part of the url
	NAME = "name"
)

//API for recipes
type API struct {
	handler core.Handler
	recipes RecipeDB
}

var (
	api *API
)

// AddRecipesAPIToHandler constructs an API for recipes
func AddRecipesAPIToHandler(handler core.Handler, recipes RecipeDB) {
	api = &API{
		handler,
		recipes,
	}

	api.prepareAPI()
}

//prepareAPI registers all api endpoints for recipes
func (rAPI *API) prepareAPI() {
	rAPI.prepareV1API()
}

func (rAPI *API) prepareV1API() {

	if rAPI.handler == nil {
		log.WithField("Component", "Recipes API").Fatal("No handler defined")
		return
	}

	if rAPI.recipes == nil {
		log.WithField("Component", "Recipes API").Fatal("No persistence defined")
		return
	}

	v1 := rAPI.handler.API(1)

	//GET the list of recipes
	v1.GET("/recipes", rAPI.getRecipes)

	//POST a new recipe
	v1.POST("/recipes", rAPI.postRecipes)

	//GET a random recipe
	v1.GET("/recipes/rand", rAPI.getRandomRecipe)

	//GET the number of recipe
	v1.GET("/recipes/num", rAPI.getNumberOfRecipes)

	//GET a specific recipe
	v1.GET("/recipes/r/:recipe", rAPI.getRecipe)

	//PUT updates a specific recipe
	v1.PUT("/recipes/r/:recipe", rAPI.putRecipe)

	//PUT updates a specific recipe
	v1.DELETE("/recipes/r/:recipe", rAPI.deleteRecipe)

	//GET a specific recipe's picture
	v1.GET("/recipes/r/:recipe/pictures/:name", rAPI.getRecipePicture)

}

// getNumberOfRecipes example
// @Summary Get the number of recipes
// @Description The number of recipes is returned that is managed by the service.
// @Tags Recipes
// @Produce json
// @Success 200 {integer} number
// @Router /recipes/num [get]
func (rAPI *API) getNumberOfRecipes(c *core.APICallContext) {
	num := rAPI.recipes.Num()
	log.Debugf("Number of Recipes %v", num)
	c.String(http.StatusOK, fmt.Sprintf("%v", num))
}

// getRecipePicture example
// @Summary Get a picture of a
// @Tags Recipes
// @Description A specific picture of a specific recipe is returned
// @Param recipe path string true "Recipe ID"
// @Param name path string true "Name of Picture"
// @Produce json
// @Success 200 {object} RecipePicture
// @Router /recipes/r/{recipe}/pictures/{name} [get]
func (rAPI *API) getRecipePicture(c *core.APICallContext) {
	recipeID := NewRecipeIDFromString(c.Param(RECIPE))
	name := c.Param(NAME)
	picture := rAPI.recipes.Picture(recipeID, name)
	if picture.ID == InvalidRecipeID() {
		c.String(http.StatusNotFound, "No such picture")
	} else {
		c.JSON(http.StatusOK, picture)
	}
}

// getRandomRecipe example
// @Summary Get a Random Recipe
// @Description A specific picture of a specific recipe is returned
// @Tags Recipes
// @Param servings query int false "Number of Servings"
// @Produce json
// @Success 200 {object} Recipe
// @Router /recipes/rand [get]
func (rAPI *API) getRandomRecipe(c *core.APICallContext) {
	query := c.Request.URL.Query()
	servings := extractServings(query)

	recipe := rAPI.recipes.Random()

	if servings > 0 {
		recipe.ScaleTo(servings)
	}

	if recipe.ID == InvalidRecipeID() {
		c.String(http.StatusNotFound, "No such recipe")
	} else {
		c.JSON(http.StatusOK, recipe)
	}
}

// getRecipes example
// @Summary Get Recipes
// @Description A list of ids of recipes is returned
// @Tags Recipes
// @Produce json
// @Success 200 {object} []string
// @Router /recipes [get]
func (rAPI *API) getRecipes(c *core.APICallContext) {
	c.JSON(http.StatusOK, rAPI.recipes.IDs())
}

// getRecipe example
// @Summary Get a specific Recipe
// @Description A specific recipe is returned
// @Tags Recipes
// @Param servings query int false "Number of Servings"
// @Param recipe path string true "Recipe ID"
// @Produce json
// @Success 200 {object} Recipe
// @Router /recipes/r/{recipe} [get]
func (rAPI *API) getRecipe(c *core.APICallContext) {
	recipeIDS := c.Param(RECIPE)
	recipeID := NewRecipeIDFromString(recipeIDS)

	query := c.Request.URL.Query()
	servings := extractServings(query)

	recipe := rAPI.recipes.Get(recipeID)

	if servings > 0 {
		recipe.ScaleTo(servings)
	}

	if recipe.ID == InvalidRecipeID() {
		c.String(http.StatusNotFound, "No such recipe: %v", recipeIDS)
	} else {
		c.JSON(http.StatusOK, recipe)
	}
}

// putRecipe example
// @Summary Update a specific Recipe
// @Description A specific recipe is updates
// @Tags Recipes
// @Param recipe path string true "Recipe ID"
// @Param message body Recipe true "Recipe"
// @Accept json
// @Produce json
// @Success 200
// @Router /recipes/r/{recipe} [put]
func (rAPI *API) putRecipe(c *core.APICallContext) {

	recipeIDS := c.Param(RECIPE)
	recipeID := NewRecipeIDFromString(recipeIDS)

	log.Error("Put Recipes called")

	var recipe Recipe
	err := c.BindJSON(&recipe)
	if err != nil || rAPI.recipes.Get(recipeID).ID == InvalidRecipeID() {
		c.String(http.StatusBadRequest, "Could not read JSON input")
	} else {
		recipe.ID = recipeID
		err = rAPI.recipes.Update(recipeID, &recipe)
		if err != nil {
			c.String(http.StatusInternalServerError, "Could not persist Recipe")
		} else {
			c.Status(http.StatusNoContent)
		}
	}
}

// postRecipes example
// @Summary Add a new Recipe
// @Description Adds a new recipe, the id will automatically overriden by the backend
// @Tags Recipes
// @Param message body Recipe true "Recipe"
// @Accept json
// @Produce json
// @Success 201
// @Router /recipes [post]
func (rAPI *API) postRecipes(c *core.APICallContext) {
	var recipe Recipe
	err := c.BindJSON(&recipe)
	if err != nil {
		c.String(http.StatusBadRequest, "Could not read JSON input")
	} else {
		recipe.ID = NewRecipeID()
		err = rAPI.recipes.Insert(&recipe)
		if err != nil {
			c.String(http.StatusInternalServerError, "Could not persist Recipe")
		} else {
			c.Status(http.StatusCreated)
		}
	}
}

// deleteRecipe example
// @Summary Delete a Recipe
// @Description Deletes a recipe by id
// @Tags Recipes
// @Param recipe path string true "Recipe ID"
// @Accept json
// @Produce json
// @Success 200
// @Router /recipes/r/{recipe} [delete]
func (rAPI *API) deleteRecipe(c *core.APICallContext) {
	recipeIDS := c.Param(RECIPE)
	recipeID := NewRecipeIDFromString(recipeIDS)
	if err := rAPI.recipes.RemoveByID(recipeID); err != nil {
		c.String(http.StatusNotFound, "Recipe not found")
		log.WithError(err).Debug("Could not Delete Recipe")
	} else {
		c.Status(http.StatusOK)
	}
}

func extractServings(query url.Values) int {
	servings := -1
	if len(query[SERVINGS]) > 0 {
		servingsS := query[SERVINGS][0]
		if num, err := strconv.Atoi(servingsS); err == nil {
			servings = num
		} else {
			log.WithError(err).Error("Could not convert the amount of servings requested")
		}
	}
	return servings
}
