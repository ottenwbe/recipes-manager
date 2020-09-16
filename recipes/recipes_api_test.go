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
	"github.com/ottenwbe/go-cook/core"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("recipesAPI", func() {

	var (
		server core.Server
	)

	BeforeSuite(func() {
		handler := core.NewHandler()
		recipes, _ := NewDatabaseClient()
		AddRecipesAPIToHandler(handler, recipes)
		server = core.NewServerA(":8080", handler)
		server.Run()
		time.Sleep(500 * time.Millisecond)
	})

	AfterSuite(func() {
		server.Close()
	})

	Context("Creating the API V1", func() {
		It("should get created", func() {
			resp, err := http.Get("http://localhost:8080/api/v1/recipes")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Context("Getting Recipes", func() {
		It("random recipe with empty get created", func() {
			resp, err := http.Get("http://localhost:8080/api/v1/recipes/rand")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))
		})
	})

	Context("Getting Recipes", func() {
		It("random recipe with empty get created", func() {
			resp, err := http.Get("http://localhost:8080/api/v1/recipes/num")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

})
