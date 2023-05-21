package account_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ottenwbe/recipes-manager/account"
)

var _ = Describe("Account DB", func() {

	Context("State", func() {
		It("can be stored and be found", func() {
			state := account.States.CreateState("test-url", true)
			result, ok := account.States.Get(state.StateString)
			Expect(ok).To(BeTrue())
			Expect(result).To(Equal(state))
		})

		It("can be found and deleted", func() {
			state := account.States.CreateState("test-url", true)

			result := account.States.FindAndDelete(state.StateString)
			_, ok := account.States.Get(state.StateString)

			Expect(result).ToNot(BeNil())
			Expect(ok).To(BeFalse())
		})
	})

})
