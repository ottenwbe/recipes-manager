/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRecipes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Recipes Suite")
}
