/*
 * MIT License - see LICENSE file for details
 */

package sources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DriveClient", func() {
	Context("OpenNewGoogleDriveConnection", func() {

		var (
			client = OpenNewGoogleDriveConnection()
		)

		It("should create a valid DriveClient", func() {
			Expect(client).ToNot(BeNil())
		})

	})
})
