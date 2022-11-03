package smt

import "fmt"

// MapDb is a key-value storage like a Database.
type MapDb interface {
	Get(key []byte) ([]byte, error)     // Get gets the value for a key.
	Set(key []byte, value []byte) error // Set updates the value for a key.
	Delete(key []byte) error            // Delete deletes a key.
}

// InvalidKey is thrown when a key that does not exist is being accessed.
type InvalidKey struct {
	Key []byte
}

func (e *InvalidKey) Error() string {
	return fmt.Sprintf("invalid key: %x", e.Key)
}

// Map is a simple in-memory map.
type Map struct {
	m map[string][]byte
}

//NewMap creates a new empty SimpleMap.
func NewMap() *Map {
	return &Map{
		m: make(map[string][]byte),
	}
}

// Get gets the value for a key.
func (sm *Map) Get(key []byte) ([]byte, error) {
	if value, ok := sm.m[string(key)]; ok {
		return value, nil
	}
	return nil, &InvalidKey{Key: key}
}

// Set updates the value for a key.
func (sm *Map) Set(key []byte, value []byte) error {
	sm.m[string(key)] = value
	return nil
}

// Delete deletes a key.
func (sm *Map) Delete(key []byte) error {
	_, ok := sm.m[string(key)]
	if ok {
		delete(sm.m, string(key))
		return nil
	}
	return &InvalidKey{Key: key}
}
