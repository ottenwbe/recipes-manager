/*
 * MIT License - see LICENSE file for details
 */

package sources

import (
	"fmt"
	"io"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/ottenwbe/recipes-manager/config"
	"github.com/ottenwbe/recipes-manager/recipes"
)

var _ = Describe("drive parser", func() {

	var (
		recipeHTML          io.Reader
		expectedIngredients = []recipes.Ingredients{
			{Name: "Something", Amount: 150, Unit: "g"},
			{Name: "Others", Amount: 2, Unit: "parts"},
			{Name: "Flour", Amount: -1, Unit: ""},
		}
		expectedID     = recipes.NewRecipeID()
		expectedRecipe = &recipes.Recipe{
			ID:          expectedID,
			Name:        "TestRecipe",
			Ingredients: expectedIngredients,
			Description: "\nDo sth.\n\nAnd sth else.\n\nAnd some salad.",
			PictureLink: []string{"IMG_20141227_132212.jpg"},
			Servings:    1,
		}
	)

	const recipeFile = "fixtures/default-recipe.html"

	BeforeEach(func() {
		config.Config.SetDefault(DriveParserIngredientsTitle, "Zutaten")
		config.Config.SetDefault(DriveRecipeInstructionsTitle, "Zubereitung")

		var err error
		recipeHTML, err = os.Open(recipeFile)
		if err != nil {
			log.WithError(err).Println("Error reading file ...")
			Fail(fmt.Sprintf("Error reading fixtures/default-recipe.html %v", recipeFile))
		}
	})

	It("parses the default recipe's html and returns a valid *Recipe", func() {
		recipe, _, err := ParseRecipe(recipeHTML, expectedID)
		Expect(err).To(BeNil())
		Expect(recipe).To(Equal(expectedRecipe))
		//Expect(pictures).To(Equal(expectedPicture))
	})

	It("returns an error when no valid recipe is present in html", func() {
		_, _, err := ParseRecipe(strings.NewReader("<html>"), expectedID)
		Expect(err).ToNot(BeNil())
	})
})
