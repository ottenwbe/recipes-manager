package account

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

// State stored when getting Token
type State struct {
	URL         string
	Signup      bool
	StateString string
}

// StateMap type to cache all states
type StateMap struct {
	sync.Mutex
	m map[string]*State
}

// States cached for reuse
var States *StateMap

func init() {
	States = &StateMap{m: make(map[string]*State)}
}

// Get the state by key
func (sm *StateMap) Get(key string) (*State, bool) {
	sm.Lock()
	defer sm.Unlock()
	result, ok := sm.m[key]
	return result, ok
}

// FindAndDelete returns the state for a given key and deletes the entry of the StateMap
func (sm *StateMap) FindAndDelete(key string) *State {
	sm.Lock()
	defer sm.Unlock()
	if result, ok := sm.m[key]; ok {
		delete(sm.m, key)
		return result
	}
	return nil
}

// CreateState to reuse later
func (sm *StateMap) CreateState(url string, signup bool) *State {
	var stateSeed uint64
	err := binary.Read(rand.Reader, binary.LittleEndian, &stateSeed)
	if err != nil {
		log.Error(err.Error())
	}
	stateString := fmt.Sprintf("%x", stateSeed)

	state := newState(url, signup, stateString)

	v, _ := json.Marshal(state)
	log.Infof("State Stored: %s", string(v))

	sm.Lock()
	defer sm.Unlock()
	sm.m[stateString] = state
	return state
}

func newState(url string, signup bool, stateString string) *State {
	return &State{
		StateString: stateString,
		URL:         url,
		Signup:      signup,
	}
}
