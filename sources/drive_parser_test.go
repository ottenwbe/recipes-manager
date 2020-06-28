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

package sources

import (
	"fmt"
	"io"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/ottenwbe/go-cook/recipes"
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
		}
	)

	const recipeFile = "fixtures/default-recipe.html"

	BeforeEach(func() {

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
