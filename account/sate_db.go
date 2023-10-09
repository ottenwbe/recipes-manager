package account

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ottenwbe/recipes-manager/core"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

const STATE = "state"

// State stored when getting Token
type State struct {
	URL       string `json:"url"`
	Signup    bool   `json:"signup"`
	State     string `json:"state"`
	CreatedAt int64  `json:"created_at"`
}

// StateService type to interact with (cached) states for authentication
type StateService struct {
	dbClient *core.MongoClient
}

var (
	once         sync.Once
	stateService *StateService
)

func NewStateService(db core.DB) *StateService {

	if (db != nil) && (stateService == nil) {
		once.Do(func() {
			stateService = &StateService{dbClient: db.(*core.MongoClient)}

			err := stateService.createTextIndex()
			if err != nil {
				logrus.Error("Error creating Status DB Index", err)
			}
		})
	}

	return stateService
}

func (sm *StateService) getStatusCollection() *mongo.Collection {
	return sm.dbClient.Client.Database("accounts").Collection("status")
}

func (*StateService) ctx() context.Context {
	return context.Background()
}

func (sm *StateService) createTextIndex() interface{} {
	c := sm.getStatusCollection()

	textIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: STATE, Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := c.Indexes().CreateOne(sm.ctx(), textIndex)

	return err
}

// Get the state by key
func (sm *StateService) Get(key string) (*State, error) {

	collection := sm.getStatusCollection()
	var result State

	err := collection.FindOne(sm.ctx(), bson.M{STATE: key}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FindAndDelete returns the state for a given key and deletes the entry of the StateMap
func (sm *StateService) FindAndDelete(key string) *State {

	if result, err := sm.Get(key); err == nil {
		collection := sm.getStatusCollection()
		_, err := collection.DeleteOne(sm.ctx(), bson.M{STATE: result.State})
		if err != nil {
			log.Error(err)
		}
		return result
	} else {
		log.Error(err)
	}
	return nil
}

// CreateState to reuse later
func (sm *StateService) CreateState(url string, signup bool) *State {
	var stateSeed uint64

	err := binary.Read(rand.Reader, binary.LittleEndian, &stateSeed)
	if err != nil {
		log.Error(err.Error())
	}
	stateString := fmt.Sprintf("%x", stateSeed)

	state := newState(url, signup, stateString)

	v, _ := json.Marshal(state)
	log.Infof("State Stored: %s", string(v))

	collection := sm.getStatusCollection()
	_, err = collection.InsertOne(stateService.ctx(), state)
	if err != nil {
		log.Error("Could not store state", err)
	}

	return state
}

func newState(url string, signup bool, stateString string) *State {
	return &State{
		State:     stateString,
		URL:       url,
		Signup:    signup,
		CreatedAt: time.Now().UnixNano(), //TODO: need to delete outdated entries
	}
}
