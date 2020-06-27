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

package core

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {

	Context("Struct", func() {
		It("can be created", func() {
			v := Version{API: "vApi", App: "vApp"}
			Expect(v.API).To(Equal("vApi"))
			Expect(v.App).To(Equal("vApp"))
		})

		It("can be marshaled to json", func() {
			v := Version{API: "vApi", App: "vApp"}
			b, err := json.Marshal(v)
			expected := "{\"app\":\"vApp\",\"api\":\"vApi\"}"
			Expect(err).To(BeNil())
			Expect(string(b)).To(Equal(expected))
		})
	})

	Context("Getter", func() {
		It("the result exactly once", func() {
			appVersionString = "test1"
			v1 := AppVersion()
			appVersionString = "test2"
			v2 := AppVersion()
			Expect(v1).To(Equal(v2))
		})
	})
})
