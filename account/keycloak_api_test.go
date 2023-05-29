package account_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/ottenwbe/recipes-manager/core"
)

var (
	db     core.DB
	server core.Server
)

/*var _ = BeforeSuite(func() {
	handler := core.NewHandler()
	db, _ = core.NewDatabaseClient()
	account.AddAuthAPIsToHandler(handler, db)
	server = core.NewServerA(":8090", handler)
	server.Run()
	time.Sleep(500 * time.Millisecond)
})*/

//var _ = AfterSuite(func() {
//	err := server.Close()
//	if err != nil {
//		Fail(err.Error())
//	}
//	err = db.Close()
//	if err != nil {
//		Fail(err.Error())
//	}
//})

var _ = Describe("Keycloak API", func() {

})
