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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("recipes", func() {

	Context("id", func() {
		It("should be able to produce a consistent invalid id", func() {
			id1 := InvalidRecipeID()
			id2 := InvalidRecipeID()
			Expect(id1).To(Equal(id2))
		})

		It("should produce different ids with each call to NewRecipeID", func() {
			id1 := NewRecipeID()
			id2 := NewRecipeID
			Expect(id1).ToNot(Equal(id2()))
		})

		It("should be able to convert an id from and to a string", func() {
			idString := "40ac4297-d5b3-435e-9f42-e5e3479d0ae8"
			id := NewRecipeIDFromString(idString)
			Expect(id.String()).To(Equal(idString))
		})

		It("should return an invalid id when it cannot convert a string to a uuid", func() {
			idFromString := NewRecipeIDFromString("inv")
			Expect(idFromString).To(Equal(InvalidRecipeID()))
		})

	})

	Context("creation", func() {
		It("should be able to create a recipe with a given id", func() {
			expectedID := NewRecipeID()
			retrieved := NewRecipe(expectedID)
			Expect(retrieved.ID).To(Equal(expectedID))
			Expect(retrieved.PictureLink).NotTo(BeNil())
		})
		It("should be able to create an invalid recipe with an invalid id", func() {
			expectedID := InvalidRecipeID()
			retrieved := NewInvalidRecipe()
			Expect(retrieved.ID).To(Equal(expectedID))
		})
	})

	Context("conversion", func() {
		It("should be able to convert a recipe to a string", func() {
			expected := "{\"id\":\"\",\"name\":\"\",\"components\":null,\"description\":\"\",\"pictureLink\":null,\"servings\":0}"
			retrieved := &Recipe{}
			Expect(retrieved.String()).To(Equal(expected))
		})

		It("should be able to convert a recipe to a json byte string", func() {
			expected := []byte("{\"id\":\"\",\"name\":\"\",\"components\":null,\"description\":\"\",\"pictureLink\":null,\"servings\":0}")
			r := &Recipe{}
			Expect(r.JSON()).To(Equal(expected))
		})
	})

	Context("scale", func() {
		It("should be able to scale up", func() {
			recipe := Recipe{
				Servings: 1,
				Ingredients: []Ingredients{
					{Amount: 1, Name: "test1", Unit: "g"},
					{Amount: 2, Name: "test2", Unit: "l"},
				},
			}
			recipe.ScaleBy(2)
			Expect(recipe.Ingredients[0].Amount).To(Equal(2.0))
			Expect(recipe.Ingredients[1].Amount).To(Equal(4.0))
		})
		It("should be able to scale down", func() {
			recipe := Recipe{
				Servings: 1,
				Ingredients: []Ingredients{
					{Amount: 1, Name: "test1", Unit: "g"},
					{Amount: 2, Name: "test2", Unit: "l"},
				},
			}
			recipe.ScaleBy(0.5)
			Expect(recipe.Ingredients[0].Amount).To(Equal(0.5))
			Expect(recipe.Ingredients[1].Amount).To(Equal(1.0))
		})
		It("should be able to scale up to specific number of portions", func() {
			recipe := Recipe{
				Servings: 2,
				Ingredients: []Ingredients{
					{Amount: 1, Name: "test1", Unit: "g"},
					{Amount: 2, Name: "test2", Unit: "l"},
				},
			}
			recipe.ScaleTo(4)
			Expect(recipe.Ingredients[0].Amount).To(Equal(2.0))
			Expect(recipe.Ingredients[1].Amount).To(Equal(4.0))
		})
		It("should be able to scale down to specific number of portions", func() {
			recipe := Recipe{
				Servings: 2,
				Ingredients: []Ingredients{
					{Amount: 1, Name: "test1", Unit: "g"},
					{Amount: 2, Name: "test2", Unit: "l"},
				},
			}
			recipe.ScaleTo(1)
			Expect(recipe.Ingredients[0].Amount).To(Equal(0.5))
			Expect(recipe.Ingredients[1].Amount).To(Equal(1.0))
		})
		It("shouldignore invalid amounts", func() {
			recipe := Recipe{
				Servings: 2,
				Ingredients: []Ingredients{
					{Amount: -1, Name: "test1", Unit: "g"},
					{Amount: 2, Name: "test2", Unit: "l"},
				},
			}
			recipe.ScaleTo(1)
			Expect(recipe.Ingredients[0].Amount).To(Equal(-1.0))
			Expect(recipe.Ingredients[1].Amount).To(Equal(1.0))
		})
	})
})
