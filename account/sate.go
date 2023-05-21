package account

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
)

// State stored when getting Token
type State struct {
	URL         string
	Signup      bool
	StateString string
}

// StateMap type to cache all states
type StateMap map[string]*State

// States cached for reuse
var States StateMap

func init() {
	States = make(StateMap)
}

// Get the state by key
func (sm StateMap) Get(key string) (*State, bool) {
	result, ok := sm[key]
	return result, ok
}

// FindAndDelete returns the state for a given key and deletes the entry of the StateMap
func (sm StateMap) FindAndDelete(key string) *State {

	if result, ok := sm[key]; ok {
		delete(sm, key)
		return result
	}
	return nil
}

// CreateState to reuse later
func (sm StateMap) CreateState(url string, signup bool) *State {
	var stateSeed uint64
	err := binary.Read(rand.Reader, binary.LittleEndian, &stateSeed)
	if err != nil {
		log.Error(err.Error())
	}
	stateString := fmt.Sprintf("%x", stateSeed)

	state := newState(url, signup, stateString)
	sm[stateString] = state
	return state
}

func newState(url string, signup bool, stateString string) *State {
	return &State{
		StateString: stateString,
		URL:         url,
		Signup:      signup,
	}
}
