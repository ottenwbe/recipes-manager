// SPDX-License-Identifier: MIT

package account

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"net/http"
	"time"

	"github.com/ottenwbe/recipes-manager/config"
	"github.com/ottenwbe/recipes-manager/internal/db"
	"github.com/ottenwbe/recipes-manager/internal/server"
)

var (
	database  db.DB
	appServer *server.Server
	handler   server.Handler
)

var _ = Describe("Keycloak API", func() {
	BeforeEach(func() {
		handler = server.NewHandler("*")
		addr := config.Config.GetString(db.RecipeDBHostCfg)
		var err error
		database, err = db.NewDatabaseClient(addr)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := appServer.Close()
		if err != nil {
			Fail(err.Error())
		}
		if database != nil {
			err = database.Close()
			if err != nil {
				Fail(err.Error())
			}
			database = nil
		}
	})

	Context("Keycloak Disabled", func() {
		It("disables keycloak endpoints endpoints (by default)", func() {

			AddAuthAPIsToHandler(handler, database)
			appServer = server.NewServerWithAddress(":8090", handler)
			appServer.Run()
			time.Sleep(500 * time.Millisecond)

			resp, err := http.Get("http://localhost:8090/api/v1/auth/keycloak/token")

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	// Test signup
	// Test login
	// Test logout
	// Test getting a token

})
