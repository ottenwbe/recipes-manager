/*
 * MIT License  - see LICENSE file for details
 */

package sources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Context("CBytes", func() {
		It("should transform a byte array to an comma separated string", func() {
			bytes := []byte{100, 200, 50}
			Expect(CBytes(bytes)).To(Equal("[100,200,50]"))
		})
	})
})
