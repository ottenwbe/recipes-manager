/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ottenwbe/recipes-manager/core"
	log "github.com/sirupsen/logrus"
)

const (
	// SERVINGS keyword used as part of the url
	SERVINGS = "servings"
	// RECIPE keyword used as part of the url
	RECIPE = "recipe"
	// NAME keyword used as part of the url
	NAME = "name"
	// INGREDIENT keyword used as part of the url
	INGREDIENT = "ingredient"
	// DESCRIPTION keyword used as part of the url
	DESCRIPTION = "description"
)

// API for recipes
type API struct {
	handler core.Handler
	recipes RecipeDB
}

// AddRecipesAPIToHandler constructs an API for recipes
func AddRecipesAPIToHandler(handler core.Handler, recipesDB RecipeDB) error {

	api := &API{
		handler,
		recipesDB,
	}

	api.prepareAPI()

	return nil
}

// prepareAPI registers all api endpoints for recipes
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

	// GET the list of recipes
	v1.GET("/recipes", rAPI.getRecipes)

	// POST a new recipe
	v1.POST("/recipes", rAPI.postRecipes)

	// GET a random recipe
	v1.GET("/recipes/rand", rAPI.getRandomRecipe)

	// GET the number of recipe
	v1.GET("/recipes/num", rAPI.getNumberOfRecipes)

	// GET a specific recipe
	v1.GET("/recipes/r/:recipe", rAPI.getRecipe)

	// PUT updates a specific recipe
	v1.PUT("/recipes/r/:recipe", rAPI.putRecipe)

	// PUT updates a specific recipe
	v1.DELETE("/recipes/r/:recipe", rAPI.deleteRecipe)

	// GET a specific recipe's picture
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
		c.JSON(http.StatusNotFound, core.H{"error": "picture not found", "id": recipeID, "name": name})
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
		c.JSON(http.StatusNotFound, core.H{"error": "no recipes found"})
	} else {
		c.JSON(http.StatusOK, recipe)
	}
}

// getRecipes example
// @Summary Get Recipes
// @Description A list of ids of recipes is returned
// @Tags Recipes
// @Param name query string false "Search for a specific name"
// @Param description query string false "Search for a specific term in a description"
// @Param ingredient query string false "Search for a specific ingredient"
// @Produce json
// @Success 200 {object} RecipeList
// @Router /recipes [get]
func (rAPI *API) getRecipes(c *core.APICallContext) {

	query := c.Request.URL.Query()

	searchFilter := extractSearchFilter(query)

	debugFilterJSON, _ := json.Marshal(searchFilter)
	log.WithField("json", string(debugFilterJSON)).Debug("Get Recipes")

	c.JSON(http.StatusOK, rAPI.recipes.IDs(searchFilter))
}

// getRecipe documentation
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
		c.JSON(http.StatusNotFound, core.H{"error": "recipe not found", "id": recipeIDS})
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
// @Success 204
// @Router /recipes/r/{recipe} [put]
func (rAPI *API) putRecipe(c *core.APICallContext) {

	recipeID := NewRecipeIDFromString(c.Param(RECIPE))
	log.WithField("id", recipeID).Debug("Put Recipe called")

	// 1. Check if recipe is valid
	if rAPI.recipes.Get(recipeID).ID == InvalidRecipeID() {
		c.JSON(http.StatusNotFound, core.H{"error": "recipe not found", "id": recipeID})
		return
	}

	// 2. Bind the request body
	var recipe Recipe
	if err := c.BindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, core.H{"error": "invalid json format"})
		return
	}

	// 3. Perform the update
	recipe.ID = recipeID // Ensure the ID from the URL is used, not from the body
	if err := rAPI.recipes.Update(recipeID, &recipe); err != nil {
		c.JSON(http.StatusInternalServerError, core.H{"error": "could not update recipe"})
	} else {
		c.Status(http.StatusNoContent)
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
	if err := c.BindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, core.H{"error": "invalid json format"})
		return
	}

	recipe.ID = NewRecipeID() // Assign a new ID, ignoring any client-provided ID
	if err := rAPI.recipes.Insert(&recipe); err != nil {
		c.JSON(http.StatusInternalServerError, core.H{"error": "could not persist recipe"})
	} else {
		c.Status(http.StatusCreated)
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
// @Success 204
// @Router /recipes/r/{recipe} [delete]
func (rAPI *API) deleteRecipe(c *core.APICallContext) {
	recipeID := NewRecipeIDFromString(c.Param(RECIPE))
	if err := rAPI.recipes.Remove(recipeID); err != nil {
		// Assuming the DB layer returns an error when the item is not found.
		c.JSON(http.StatusNotFound, core.H{"error": "recipe not found", "id": recipeID})
		log.WithError(err).Debug("Could not Delete Recipe")
	} else {
		c.Status(http.StatusNoContent) // 204 is more idiomatic for a successful DELETE with no body
	}
}

func extractServings(query url.Values) int8 {
	var servings int64 = -1
	if len(query[SERVINGS]) > 0 {
		servingsS := query[SERVINGS][0]
		if num, err := strconv.ParseInt(servingsS, 10, 8); err == nil {
			servings = num
		} else {
			log.WithError(err).Error("Could not convert the amount of servings requested")
		}
	}
	return int8(servings)
}

func extractSearchString(query url.Values, param string) string {
	var result = ""

	if len(query[param]) > 0 {
		result = query[param][0]
	}

	return result
}

func extractIngredientSearchArray(query url.Values) []string {
	return query[INGREDIENT]
}

func extractSearchFilter(query url.Values) *RecipeSearchFilter {
	return &RecipeSearchFilter{
		Ingredient:  extractIngredientSearchArray(query),
		Name:        extractSearchString(query, NAME),
		Description: extractSearchString(query, DESCRIPTION),
	}
}
