/*
 * MIT License
 *
 * Copyright (c) 2020 Beate Ottenwälder
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package sources

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"

	. "github.com/ottenwbe/recipes-manager/utils"
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

			id := uuid.NewV4()
			meta := NewSourceDescription(SourceID(id), "name", "test", nil)
			expected := fmt.Sprintf("{\"id\":%v,\"name\":\"name\",\"connected\":true,\"version\":\"test\"}", CBytes(id.Bytes()))

			b, err := json.Marshal(meta)

			Expect(err).To(BeNil())
			Expect(string(b)).To(Equal(expected))
		})
	})

})
