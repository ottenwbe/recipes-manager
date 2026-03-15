/*
 * MIT License - see LICENSE file for details
 */

package recipes

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Context("UniqueSlice", func() {
		It("should remove duplicates from a slice of strings", func() {
			input := []string{"a", "b", "a", "c", "b"}
			expected := []string{"a", "b", "c"}
			Expect(UniqueSlice(input)).To(Equal(expected))
		})

		It("should return an empty slice when input is empty", func() {
			input := []string{}
			Expect(UniqueSlice(input)).To(BeEmpty())
		})
	})
})
