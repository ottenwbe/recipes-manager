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

	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

//Ingredients of a recipe
type Ingredients struct {
	//Name of the ingredient
	Name string `json:"name"`
	//Amount needed in a recipe of an ingredient
	Amount float64 `json:"amount"`
	//Unit of the Amount
	Unit string `json:"unit"`
}

const (
	//NoAmountIngredient is the amount value for ingredients when this field is not used
	NoAmountIngredient = -1.0
)

//RecipeID is a data type that provides a unique id for each recipe
type RecipeID string

//String converts a RecipeID to string
func (r RecipeID) String() string {
	return string(r)
}

//InvalidRecipeID should not be used for any valid Recipe
func InvalidRecipeID() RecipeID {
	return RecipeID(uuid.Nil.String())
}

//NewRecipeID returns a random recipe id
func NewRecipeID() RecipeID {
	return RecipeID(uuid.NewV4().String())
}

//NewRecipeIDFromString converts a string to a recipe id and returns this recipe id.
//Returns the InvalidRecipeID iff the recipe id cannot be converted
func NewRecipeIDFromString(recipeID string) (result RecipeID) {
	if tmp, err := uuid.FromString(recipeID); err != nil {
		result = InvalidRecipeID()
	} else {
		result = RecipeID(tmp.String())
	}
	return
}

//Recipe model
type Recipe struct {
	ID          RecipeID      `json:"id"`
	Name        string        `json:"name"`
	Ingredients []Ingredients `json:"components"`
	Description string        `json:"description"`
	PictureLink []string      `json:"pictureLink"`
	Servings    int8          `json:"servings"`
}

//RecipePicture model
type RecipePicture struct {
	ID      RecipeID `json:"id"`
	Name    string   `json:"name"`
	Picture string   `json:"picture"`
}

//NewInvalidRecipePicture returns an invalid picture
func NewInvalidRecipePicture() *RecipePicture {
	return &RecipePicture{
		ID:      InvalidRecipeID(),
		Name:    "",
		Picture: "",
	}
}

//RecipeSearchFilter models a search query to filter recipes
type RecipeSearchFilter struct {
	Name        string   `json:"name"`
	Ingredient  []string `json:"ingredients"`
	Description string   `json:"description"`
}

//Recipes interface is an abstraction for the provider of a collection of recipes, i.e., a data-base or a cache
type Recipes interface {
	List() []*Recipe
	IDs(filterQuery *RecipeSearchFilter) []string
	Num() int64
	Get(id RecipeID) *Recipe
	GetByName(name string) (*Recipe, error)
	Picture(id RecipeID, name string) *RecipePicture
	Pictures(id RecipeID) map[string]*RecipePicture
	Random() *Recipe
	Insert(recipe *Recipe) error
	Update(id RecipeID, recipe *Recipe) error
	AddPicture(pic *RecipePicture) error
	Remove(id RecipeID) error
	RemoveByName(name string) error
}

const (
	//NotSupportedError informs clients about operations that are not supported by a Recipes provider
	NotSupportedError = "recipe operation not supported"
)

//NewInvalidRecipe returns an empty Recipe object. The ID of the returned Recipe is InvalidRecipeID.
func NewInvalidRecipe() *Recipe {
	return &Recipe{
		ID:   InvalidRecipeID(),
		Name: "No Recipe",
	}
}

//NewRecipe creates a new Recipe with a given id
func NewRecipe(id RecipeID) *Recipe {
	return &Recipe{
		ID:          id,
		Name:        "",
		Ingredients: make([]Ingredients, 0),
		Description: "",
		PictureLink: make([]string, 0),
		Servings:    1,
	}
}

//JSON returns the encoded version of the recipe. If an error occurs, '{}' is returned.
func (r *Recipe) JSON() []byte {
	bytes, err := json.Marshal(r)
	if err != nil {
		log.WithError(err).Error("Could not convert recipe to bytes!")
		return []byte("{}")
	}
	return bytes
}

//String (JSON) representation of the recipe
func (r *Recipe) String() string {
	return string(r.JSON())
}

//ScaleBy a factor (of servings) all ingredients of the recipe
func (r *Recipe) ScaleBy(factor float64) {
	for i := range r.Ingredients {
		if r.Ingredients[i].Amount > 0 {
			r.Ingredients[i].Amount *= factor
		}
	}
}

//ScaleTo a desired number of servings
func (r *Recipe) ScaleTo(servings int8) {
	factor := float64(servings) / float64(r.Servings)
	r.Servings = servings
	r.ScaleBy(factor)
}
