// SPDX-License-Identifier: MIT

package version

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {

	Context("Struct", func() {
		It("can be created", func() {
			v := Version{API: "vApi", App: "vApp"}
			Expect(v.API).To(Equal("vApi"))
			Expect(v.App).To(Equal("vApp"))
		})

		It("can be marshaled to json", func() {
			v := Version{API: "vApi", App: "vApp"}
			b, err := json.Marshal(v)
			expected := "{\"app\":\"vApp\",\"api\":\"vApi\"}"
			Expect(err).To(BeNil())
			Expect(string(b)).To(Equal(expected))
		})
	})

	Context("Getter", func() {
		It("the result exactly once", func() {
			appVersionString = "test1"
			v1 := AppVersion()
			appVersionString = "test2"
			v2 := AppVersion()
			Expect(v1).To(Equal(v2))
		})
	})
})
