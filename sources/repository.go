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
	"errors"

	log "github.com/sirupsen/logrus"
)

// DefaultSources is the default implementation of a sourceClient repository
type DefaultSources struct {
	sources map[SourceID]*Source
}

// NewSources creates and returns a new sourceClient repository
func NewSources() Sources {
	return &DefaultSources{
		sources: make(map[SourceID]*Source),
	}
}

// JSON returns all sources as JSON
func (s *DefaultSources) JSON() ([]byte, error) {
	result := map[string]*SourceDescription{}

	for k, v := range s.sources {
		result[k.String()] = v.sourceDescription
	}

	return json.Marshal(result)
}

// List all sources and the corresponding sourceDescription information
func (s *DefaultSources) List() (map[SourceID]*SourceDescription, error) {
	log.Debugf("list %v", s.sources)

	if len(s.sources) > 0 {
		return s.descriptionMap(), nil
	}

	return nil, errors.New("empty List of Sources")
}

func (s *DefaultSources) descriptionMap() map[SourceID]*SourceDescription {
	result := map[SourceID]*SourceDescription{}

	for k, v := range s.sources {
		result[k] = v.sourceDescription
	}

	return result
}

// Add a sourceClient to the default sourceClient
func (s *DefaultSources) Add(sourceMeta *SourceDescription, source SourceClient) error {
	if source != nil {
		s.sources[sourceMeta.ID] = &Source{sourceMeta, source}
		return nil
	}
	return errors.New("could not add invalid sourceClient")
}

// RemoveByID will delete the element with ID id from the repository
func (s *DefaultSources) RemoveByID(id SourceID) error {
	delete(s.sources, id)
	return nil
}

// Remove will delete the given element from the repository.
// The deletion will use the sourceClient's id for this purpose.
func (s *DefaultSources) Remove(source *SourceDescription) error {
	if source != nil {
		delete(s.sources, source.ID)
		return nil
	}
	return errors.New("could not delete invalid sourceClient")
}

// Client by id returns the corresponding sourceClient implementations
func (s *DefaultSources) Client(id SourceID) (SourceClient, error) {
	if val, ok := s.sources[id]; ok {
		return val.concrete, nil
	}
	return nil, errors.New("cannot find sourceClient")
}

// Description will return a sourceClient's SourceDescription
func (s *DefaultSources) Description(id SourceID) (*SourceDescription, error) {
	if val, ok := s.sources[id]; ok {
		return val.sourceDescription, nil
	}
	return NewInvalidSourceDescription(), errors.New("cannot find sourceClient")
}
