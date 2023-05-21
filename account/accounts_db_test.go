package account_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ottenwbe/recipes-manager/account"
	"github.com/ottenwbe/recipes-manager/core"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Account DB", func() {

	var (
		database      core.DB
		mongoDatabase *account.MongoAccountService
		err           error
	)

	BeforeEach(func() {
		database, err = core.NewDatabaseClient()
		mongoDatabase = account.NewMongoAccountService(database)
	})

	AfterEach(func() {
		_ = database.Close()
	})

	Context("client", func() {
		It("can be connected to w/o an error", func() {
			Expect(err).To(BeNil())
		})
	})

	Context("can save an account", func() {

		accountName := "test_create_account"

		It("while not throwing an error", func() {
			acc, err := mongoDatabase.NewAccount(accountName)

			Expect(err).To(BeNil())
			Expect(acc.Name).To(Equal(accountName))
		})

		It("cannot create an entry twice", func() {
			_, _ = mongoDatabase.NewAccount(accountName)
			_, err = mongoDatabase.NewAccount(accountName)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("can find an account", func() {

		accountName := "test_find_acc"

		It("that has been stored beforehand", func() {
			a, err := mongoDatabase.NewAccount(accountName)

			logrus.Info(json.Marshal(a))

			bAcc, err := mongoDatabase.FindAccount(accountName)

			Expect(err).To(BeNil())
			Expect(bAcc.Name).To(Equal(accountName))
			Expect(bAcc.ID).To(Equal(a.ID))
		})
	})

	Context("can delete", func() {

		accountName := "test_del_acc"

		It("a stored account by a ID", func() {
			idAccount, err := mongoDatabase.NewAccount(accountName)
			err = mongoDatabase.DeleteAccountByID(idAccount.ID)
			foundAcc, err := mongoDatabase.FindAccount(accountName)

			Expect(err).ToNot(BeNil())
			Expect(foundAcc).To(BeNil())
		})

		It("a stored account by a name", func() {
			_, err = mongoDatabase.NewAccount(accountName)
			err = mongoDatabase.DeleteAccountByName(accountName)
			foundAcc, err := mongoDatabase.FindAccount(accountName)

			Expect(err).ToNot(BeNil())
			Expect(foundAcc).To(BeNil())
		})
	})
})
