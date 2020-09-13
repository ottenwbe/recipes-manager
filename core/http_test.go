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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"reflect"

	"github.com/ottenwbe/go-cook/utils"
)

var _ = Describe("http", func() {

	Context("configuration of the http components", func() {
		It("should use the default ADDRESS if no address is given", func() {
			Expect(utils.Config.GetString(addressCfg)).To(Equal(":8080"))
		})
	})

	Context("creation of the http router", func() {
		It("should be of type Router", func() {
			r := NewRouter()
			Expect(reflect.TypeOf(r)).To(Equal(reflect.TypeOf(&ginRouter{})))
		})
	})

	Context("creation of the server", func() {
		It("should set the configured ADDRESS", func() {
			s := NewServer()
			Expect(s.Address).To(Equal(utils.Config.GetString(addressCfg)))
		})
	})

	Context("routes", func() {
		It("should create and cache a versioned api route", func() {
			r := NewRouter()
			v1 := r.API(1)
			Expect(v1).ToNot(BeNil())
			Expect(v1.Path()).To(Equal("/api/v1"))
			Expect(r.(*ginRouter).routerGroups).To(HaveKey("v1"))
		})
	})
})
