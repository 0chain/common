package statecheck

import (
	"errors"
	"reflect"
)

// StateCheck is a helper tool that should only be used in development environment.
// The main purpose of this tool is to check if there are any changes in the state
// that are not reflected in the MPT.
//
// How to manage the nodes that acquired from the state checker, and then re-copied and used
// in SC. In that case, any changes to that copied node will be reflected in the state checker.
//
// Options are:
// 1. Update the node acquired from the state checker with functions/methods provided by the
// state checker. In this way, we can make sure that all changes are reflected, well the disadvantage
// is that we need to refact all the places to use the functions/methods, which is a lot of work.
// 2. Thinking about other options.

// type StateChecker interface {
// 	Add(key []byte, value interface{}) error
// 	Get(key []byte) (interface{}, error)
// }

// StateCheck is a state checker
type StateCheck struct {
	stateNodes map[string]interface{}
}

// NewStateCheck creates a new state checker
func NewStateCheck() *StateCheck {
	return &StateCheck{
		stateNodes: make(map[string]interface{}),
	}
}

// Add adds a new key-value pair to the state checker
func (sc *StateCheck) Add(key []byte, value interface{}) error {
	// the value must be a pointer
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		return errors.New("value must be a pointer")
	}

	_, ok := sc.stateNodes[string(key)]
	if ok {
		return errors.New("key already exists")
	}

	sc.stateNodes[string(key)] = value
	return nil
}

// Get returns the value associated with the key
func (sc *StateCheck) Get(key []byte) (interface{}, error) {
	value, ok := sc.stateNodes[string(key)]
	if !ok {
		return nil, errors.New("key not found")
	}

	return value, nil
}

// ForEach iterates over all the keys in the state checker
func (sc *StateCheck) ForEach(handler func(key []byte, value interface{}) error) error {
	for k, v := range sc.stateNodes {
		if err := handler([]byte(k), v); err != nil {
			return err
		}
	}
	return nil
}
