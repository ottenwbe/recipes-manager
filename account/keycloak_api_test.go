package account

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ottenwbe/recipes-manager/core"
	"net/http"
	"time"
)

var (
	db      core.DB
	server  core.Server
	handler core.Handler
)

var _ = Describe("Keycloak API", func() {
	BeforeEach(func() {
		handler = core.NewHandler()
		db, _ = core.NewDatabaseClient()
	})

	AfterEach(func() {
		err := server.Close()
		if err != nil {
			Fail(err.Error())
		}
		if db != nil {
			err = db.Close()
			if err != nil {
				Fail(err.Error())
			}
			db = nil
		}
	})

	Context("Keycloak Disabled", func() {
		It("disables keycloak endpoints endpoints (by default)", func() {

			AddAuthAPIsToHandler(handler, db)
			server = core.NewServerA(":8090", handler)
			server.Run()
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
