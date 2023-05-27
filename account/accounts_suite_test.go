package account_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestAccounts suite
func TestAccounts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Accounts Suite")
}
