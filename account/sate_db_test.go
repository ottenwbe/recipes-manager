package account_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ottenwbe/recipes-manager/account"
	"github.com/ottenwbe/recipes-manager/core"
)

var _ = Describe("State DB", func() {

	var (
		database     core.DB
		stateService *account.StateService
	)

	database, err := core.NewDatabaseClient()
	if err != nil {
		Fail("DB could not be found")
	}
	stateService = account.NewStateService(database)

	Context("State", func() {
		It("can be stored and be found", func() {
			state := stateService.CreateState("test-url", true)
			result, err := stateService.Get(state.State)
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(*state))
		})

		It("can be found and deleted", func() {
			state := stateService.CreateState("test-url-2", true)

			result := stateService.FindAndDelete(state.State)
			_, ok := stateService.Get(state.State)

			Expect(result).ToNot(BeNil())
			Expect(ok).ToNot(BeNil())
		})
	})

})
