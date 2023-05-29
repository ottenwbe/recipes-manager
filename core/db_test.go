package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ottenwbe/recipes-manager/core"
)

var _ = Describe("MongoDB", func() {
	Context("can be connected to and", func() {

		m, err := core.NewDatabaseClient()

		It("does not result in an error", func() {
			Expect(err).To(BeNil())
		})

		It("can be created and then pinged", func() {
			Expect(m.Ping()).To(BeNil())
		})

		It("cannot be created twice", func() {
			_, err := core.NewDatabaseClient()
			Expect(err).To(BeNil())
		})

	})

	Context("can be close", func() {

		m, _ := core.NewDatabaseClient()

		It("does not result in an error", func() {
			errClose := m.Close()
			errPing := m.Ping()
			Expect(errClose).To(BeNil())
			Expect(errPing).ToNot(BeNil())
		})
	})
})
