/*
 * MIT License - see LICENSE file for details
 */

package sources

import (
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/ottenwbe/recipes-manager/recipes"
)

// SourceID represents a unique id of a sourceClient
type SourceID uuid.UUID

// SourceIDFromString represents a unique id of a sourceClient
func SourceIDFromString(s string) (SourceID, error) {
	u, err := uuid.Parse(s)
	return SourceID(u), err
}

// String representation of a sourceClient's id
func (s SourceID) String() string {
	return uuid.UUID(s).String()
}

// SourceClient interface
type SourceClient interface {
	ConnectOAuth(code string) error
	Connected() bool
	Recipes() recipes.Recipes
	OAuthLoginConfig() (*oauth2.Config, error)
}

// SourceDescription describes the sourceClient in detail
type SourceDescription struct {
	ID          SourceID       `json:"id"`
	Name        string         `json:"name"`
	Connected   bool           `json:"connected"`
	Version     string         `json:"version"`
	OAuthConfig *oauth2.Config `json:"-"`
}

// NewSourceDescription is the designated way to create a SourceDescription
func NewSourceDescription(id SourceID, name string, version string, oauthConfig *oauth2.Config) *SourceDescription {
	return &SourceDescription{
		ID:          id,
		Name:        name,
		Connected:   true,
		Version:     version,
		OAuthConfig: oauthConfig,
	}
}

// NewInvalidSourceDescription returns a SourceDescription with all fields set to invalid values
func NewInvalidSourceDescription() *SourceDescription {
	return &SourceDescription{
		ID:          SourceID(uuid.Nil),
		Name:        "invalid",
		Connected:   true,
		Version:     "0.0.0",
		OAuthConfig: nil,
	}
}

// Source struct that combines SourceDescriptions with the SourceClient implementation
type Source struct {
	sourceDescription *SourceDescription
	concrete          SourceClient
}

// Sources is the interface for all sourceClient repository implementations
type Sources interface {
	JSON() ([]byte, error)
	List() (map[SourceID]*SourceDescription, error)
	Add(sourceMeta *SourceDescription, source SourceClient) error
	RemoveByID(id SourceID) error
	Remove(source *SourceDescription) error
	Description(id SourceID) (*SourceDescription, error)
	Client(id SourceID) (SourceClient, error)
}
