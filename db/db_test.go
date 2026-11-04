// SPDX-License-Identifier: MIT

package db_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ottenwbe/recipes-manager/config"
	"github.com/ottenwbe/recipes-manager/db"
)

var _ = Describe("MongoDB", func() {
	Context("can be connected to and", func() {

		addr := config.Config.GetString(db.RecipeDBHostCfg)
		m, err := db.NewDatabaseClient(addr)

		It("does not result in an error", func() {
			Expect(err).To(BeNil())
		})

		It("can be created and then pinged", func() {
			Expect(m.Ping()).To(BeNil())
		})

		It("cannot be created twice", func() {
			addr := config.Config.GetString(db.RecipeDBHostCfg)
			_, err := db.NewDatabaseClient(addr)
			Expect(err).To(BeNil())
		})

	})

	Context("can be close", func() {

		addr := config.Config.GetString(db.RecipeDBHostCfg)
		m, _ := db.NewDatabaseClient(addr)

		It("does not result in an error", func() {
			errClose := m.Close()
			errPing := m.Ping()
			Expect(errClose).To(BeNil())
			Expect(errPing).ToNot(BeNil())
		})
	})
})
