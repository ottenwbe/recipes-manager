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

package utils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	It("is initialized after starting...", func() {
		Expect(Config).ToNot(BeNil())
	})

	var (
		c RecipeConfig
	)

	BeforeEach(func() {
		recipeConfig := &viperConfig{}
		recipeConfig.initConfigFile("test-config", []string{"fixtures"})
		recipeConfig.readConfig()
		c = recipeConfig
	})

	Context("Viper Configuration", func() {
		It("can read string values from files with arbitrary name and path", func() {
			s := c.GetString("str")
			Expect(s).To(Equal("success"))
		})

		It("can read integer values from files with arbitrary name and path", func() {
			i := c.GetInt64("int")
			Expect(i).To(Equal(int64(123)))
		})

		It("can handle string default values", func() {
			const expected = "default"
			const testKey = "default-str"
			c.SetDefault(testKey, expected)
			s := c.GetString(testKey)
			Expect(s).To(Equal(expected))
		})

		It("can handle int default values", func() {
			const expected = int64(1023)
			const testKey = "default-int"
			c.SetDefault(testKey, expected)
			i := c.GetInt64(testKey)
			Expect(i).To(Equal(expected))
		})
	})
})
