// SPDX-License-Identifier: MIT

package server

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"reflect"
)

var _ = Describe("http", func() {

	Context("creation of the http handler", func() {
		It("should be of type Handler", func() {
			r := NewHandler("*")
			Expect(reflect.TypeOf(r)).To(Equal(reflect.TypeOf(&ginHandler{})))
		})
	})

	Context("creation of the server", func() {
		It("should set the configured ADDRESS", func() {
			s := NewServerWithAddress(":8080", NewHandler("*"))
			Expect(s.Address).To(Equal(":8080"))
		})
	})

	Context("routes", func() {
		It("should create and cache a versioned api route", func() {
			r := NewHandler("*")
			v1 := r.API(1)
			Expect(v1).ToNot(BeNil())
			Expect(v1.Path()).To(Equal("/api/v1"))
			Expect(r.(*ginHandler).routerGroups).To(HaveKey("v1"))
		})
	})
})
