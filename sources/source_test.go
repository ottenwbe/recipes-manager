/*
 * MIT License - see LICENSE file for details
 */

package sources

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SourceClient", func() {

	Context("SourceID", func() {
		It("can be created from string", func() {
			s := "7f1870ac-bc6b-4da6-b2b4-b3df54d671cc"
			id, err := SourceIDFromString(s)
			Expect(err).To(BeNil())
			Expect(uuid.UUID(id).String()).To(Equal(s))
		})

		It("cannot be created from an invalid string", func() {
			_, err := SourceIDFromString("invalid")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("Marshalling", func() {
		It("of a SourceDescription to json is possible", func() {

			id := uuid.New()
			meta := NewSourceDescription(SourceID(id), "name", "test", nil)
			bs, _ := id.MarshalBinary()
			expected := fmt.Sprintf("{\"id\":%v,\"name\":\"name\",\"connected\":true,\"version\":\"test\"}", CBytes(bs))

			b, err := json.Marshal(meta)

			Expect(err).To(BeNil())
			Expect(string(b)).To(Equal(expected))
		})
	})

})
