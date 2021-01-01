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

	"github.com/ottenwbe/recipes-manager/recipes"
)

//SourceID represents a unique id of a sourceClient
type SourceID uuid.UUID

//SourceIDFromString represents a unique id of a sourceClient
func SourceIDFromString(s string) (SourceID, error) {
	u, err := uuid.FromString(s)
	return SourceID(u), err
}

//String representation of a sourceClient's id
func (s SourceID) String() string {
	return uuid.UUID(s).String()
}

//SourceClient interface
type SourceClient interface {
	ConnectOAuth(code string) error
	Connected() bool
	Recipes() recipes.Recipes
	OAuthLoginConfig() (*oauth2.Config, error)
}

//SourceDescription describes the sourceClient in detail
type SourceDescription struct {
	ID          SourceID       `json:"id"`
	Name        string         `json:"name"`
	Connected   bool           `json:"connected"`
	Version     string         `json:"version"`
	OAuthConfig *oauth2.Config `json:"-"`
}

//NewSourceDescription is the designated way to create a SourceDescription
func NewSourceDescription(id SourceID, name string, version string, oauthConfig *oauth2.Config) *SourceDescription {
	return &SourceDescription{
		ID:          id,
		Name:        name,
		Connected:   true,
		Version:     version,
		OAuthConfig: oauthConfig,
	}
}

//NewInvalidSourceDescription returns a SourceDescription with all fields set to invalid values
func NewInvalidSourceDescription() *SourceDescription {
	return &SourceDescription{
		ID:          SourceID(uuid.Nil),
		Name:        "invalid",
		Connected:   true,
		Version:     "0.0.0",
		OAuthConfig: nil,
	}
}

//Source struct that combines SourceDescriptions with the SourceClient implementation
type Source struct {
	sourceDescription *SourceDescription
	concrete          SourceClient
}

//Sources is the interface for all sourceClient repository implementations
type Sources interface {
	JSON() ([]byte, error)
	List() (map[SourceID]*SourceDescription, error)
	Add(sourceMeta *SourceDescription, source SourceClient) error
	RemoveByID(id SourceID) error
	Remove(source *SourceDescription) error
	Description(id SourceID) (*SourceDescription, error)
	Client(id SourceID) (SourceClient, error)
}
