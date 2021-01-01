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
	"github.com/satori/go.uuid"
	"golang.org/x/oauth2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ottenwbe/recipes-manager/recipes"
)

var _ = Describe("sourceClient repository", func() {

	Context("Sources", func() {
		It("a sourceClient that has been added can be listed", func() {
			s := NewSources()
			var source testSource

			s.Add(NewSourceDescription(SourceID(uuid.NewV4()), "test", "0.1.0", nil), source)
			testData, err := s.List()

			Expect(err).To(BeNil())
			Expect(len(testData)).To(Equal(1))
		})

		It("a sourceClient that has been added can be retrieved", func() {
			s := NewSources()
			var (
				source         testSource
				expectedSource = NewSourceDescription(SourceID(uuid.NewV4()), "test", "0.1.0", nil)
			)

			s.Add(expectedSource, source)
			testData, err := s.Description(expectedSource.ID)

			Expect(err).To(BeNil())
			Expect(testData).To(Equal(expectedSource))
		})

		It("a sourceClient that has been added can be deleted", func() {
			s := NewSources()
			var (
				source         testSource
				expectedSource = NewSourceDescription(SourceID(uuid.NewV4()), "test", "0.1.0", nil)
			)

			s.Add(expectedSource, source)
			s.Remove(expectedSource)
			testData, err := s.Description(expectedSource.ID)

			Expect(err).ToNot(BeNil())
			Expect(testData).To(Equal(NewInvalidSourceDescription()))
		})

		It("a sourceClient that has been added can be deleted (by id)", func() {
			s := NewSources()
			var (
				source         testSource
				expectedSource = NewSourceDescription(SourceID(uuid.NewV4()), "test", "0.1.0", nil)
			)

			s.Add(expectedSource, source)
			s.RemoveByID(expectedSource.ID)
			testData, err := s.Description(expectedSource.ID)

			Expect(err).ToNot(BeNil())
			Expect(testData).To(Equal(NewInvalidSourceDescription()))
		})

	})
})

type testSource uuid.UUID

func (s testSource) OAuthLoginConfig() (*oauth2.Config, error) {
	panic("implement me")
}

func (testSource) Recipes() recipes.Recipes {
	panic("implement me")
}

func (testSource) Name() string {
	return "test sourceClient"
}

func (testSource) Version() string {
	return "1.1.1"
}

func (testSource) ConnectOAuth(code string) error {
	panic("implement me")
}

func (testSource) Connected() bool {
	panic("implement me")
}
