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
	SERVINGS = "servings"
	RECIPE   = "recipe"
	NAME     = "name"
)

//API for recipes
type API struct {
	handler core.Handler
	recipes RecipeDB
}

var (
	api *API
)

//NewRecipesAPI constructs an API for recipes
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
		log.WithField("Component","Recipes API").Fatal("No persistence defined")
		return
	}

	v1 := rAPI.handler.API(1)

	//GET the list of recipes
	v1.GET("/recipes", rAPI.getRecipes)

	//POST a new recipe
	v1.POST("/recipes", rAPI.postRecipes)

	//GET a random recipe
	v1.GET("/recipes/rand", rAPI.getRandomRecipe)

	//GET a random recipe
	v1.GET("/recipes/num", rAPI.getNumberOfRecipes)

	//GET a specific recipe
	v1.GET("/recipes/r/:recipe", rAPI.getRecipe)

	//PUT updates a specific recipe
	v1.PUT("/recipes/r/:recipe", rAPI.putRecipe)

	//GET a specific recipe's picture
	v1.GET("/recipes/r/:recipe/pictures/:name", rAPI.getRecipePicture)

}

func (rAPI *API) getNumberOfRecipes(c *core.APICallContext) {
	num := rAPI.recipes.Num()
	log.Debugf("Number of Recipes %v", num)
	c.String(http.StatusOK, fmt.Sprintf("%v", num))
}

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

func (rAPI *API) getRecipes(c *core.APICallContext) {
	c.JSON(http.StatusOK, rAPI.recipes.IDs())
}

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
			c.Status(http.StatusOK)
		}
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
