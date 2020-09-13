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
	"net/url"
	"strconv"

	"github.com/ottenwbe/go-cook/core"
	log "github.com/sirupsen/logrus"
)

//API for recipes
type API struct {
	router  core.Router
	recipes RecipeDB
}

//NewRecipesAPI constructs an API for recipes
func NewRecipesAPI(router core.Router, recipes RecipeDB) *API {
	return &API{
		router,
		recipes,
	}
}

//PrepareAPI registers all api endpoints for recipes
func (rAPI *API) PrepareAPI() {
	rAPI.prepareV1API()
}

func (rAPI *API) prepareV1API() {

	if rAPI.router == nil {
		log.Fatal("No router defined")
		return
	}

	v1 := rAPI.router.API(1)

	//GET the list of recipes
	v1.GET("/recipes", func(c *core.APICallContext) {
		c.JSON(200, rAPI.recipes.IDs())
	})

	//GET a random recipe
	v1.GET("/recipes/rand", func(c *core.APICallContext) {
		recipe := rAPI.recipes.Random()
		if recipe.ID == InvalidRecipeID() {
			c.String(404, "No such recipe")
		} else {
			c.JSON(200, recipe)
		}
	})

	//GET a random recipe
	v1.GET("/recipes/num", func(c *core.APICallContext) {
		num := rAPI.recipes.Num()
		log.Debugf("Number of Recipes %v", num)
		c.String(200, fmt.Sprintf("%v", num))
	})

	//GET a specific recipe
	v1.GET("/recipes/r/:recipe", rAPI.getRecipe)

	//GET a specific recipe's picture
	v1.GET("/recipes/r/:recipe/pictures/:name", func(c *core.APICallContext) {
		recipeID := NewRecipeIDFromString(c.Param("recipe"))
		name := c.Param("name")
		picture := rAPI.recipes.Picture(recipeID, name)
		if picture.ID == InvalidRecipeID() {
			c.String(404, "No such picture")
		} else {
			c.JSON(200, picture)
		}
	})

}

func (rAPI *API) getRecipe(c *core.APICallContext) {

	query := c.Request.URL.Query()
	portions := extractPortions(query)

	recipeIDS := c.Param("recipe")
	recipeID := NewRecipeIDFromString(recipeIDS)

	recipe := rAPI.recipes.Get(recipeID)
	recipe.ScaleTo(portions)

	if recipe.ID == InvalidRecipeID() {
		c.String(404, "No such recipe: %v", recipeIDS)
	} else {
		c.JSON(200, recipe)
	}
}

func extractPortions(query url.Values) int {
	portions := 1
	if len(query["portions"]) > 0 {
		portionsS := query["portions"][0]
		if num, err := strconv.Atoi(portionsS); err != nil {
			portions = num
		}
	}
	return portions
}
