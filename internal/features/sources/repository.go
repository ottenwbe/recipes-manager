/*
 * MIT License  - see LICENSE file for details
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
