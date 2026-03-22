/*
 * MIT License  - see LICENSE file for details
 */

package sources

import (
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ottenwbe/recipes-manager/recipes"
)

var _ = Describe("sourceClient repository", func() {

	Context("Sources", func() {
		It("a sourceClient that has been added can be listed", func() {
			s := NewSources()
			var source testSource

			s.Add(NewSourceDescription(SourceID(uuid.New()), "test", "0.1.0", nil), source)
			testData, err := s.List()

			Expect(err).To(BeNil())
			Expect(len(testData)).To(Equal(1))
		})

		It("a sourceClient that has been added can be retrieved", func() {
			s := NewSources()
			var (
				source         testSource
				expectedSource = NewSourceDescription(SourceID(uuid.New()), "test", "0.1.0", nil)
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
				expectedSource = NewSourceDescription(SourceID(uuid.New()), "test", "0.1.0", nil)
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
				expectedSource = NewSourceDescription(SourceID(uuid.New()), "test", "0.1.0", nil)
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
