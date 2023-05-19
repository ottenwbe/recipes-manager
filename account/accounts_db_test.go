package account_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ottenwbe/recipes-manager/account"
	"github.com/ottenwbe/recipes-manager/core"
)

var _ = Describe("Account DB", func() {

	var (
		database      core.DB
		mongoDatabase *account.MongoAccountDB
		err           error
	)

	BeforeEach(func() {
		database, err = core.NewDatabaseClient()
		mongoDatabase = account.NewMongoAccountClient(database)
	})

	AfterEach(func() {
		_ = database.Close()
	})

	It("can be connected to w/o an error", func() {
		Expect(err).To(BeNil())
	})

	Context("can save an account", func() {

		createAccount := account.NewAccount("test")

		It("while Not throwing an error", func() {
			err = mongoDatabase.SaveAccount(createAccount)

			Expect(err).To(BeNil())
		})
	})

	Context("can find an account", func() {

		a := account.NewAccount("test_find_acc")

		It("that has been stored beforehand", func() {
			err = mongoDatabase.SaveAccount(a)
			bAcc, err := mongoDatabase.FindAccount(a)

			Expect(err).To(BeNil())
			Expect(bAcc.Name).To(Equal(a.Name))
		})
	})

	Context("can delete", func() {

		a := account.NewAccount("test_del_acc")

		It("an account", func() {
			err = mongoDatabase.SaveAccount(a)
			err = mongoDatabase.DeleteAccount(a)
			foundAcc, err := mongoDatabase.FindAccount(a)

			Expect(err).ToNot(BeNil())
			Expect(foundAcc).To(BeNil())
		})
	})
})
