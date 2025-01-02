/*
Copyright 2024 The KubeService-Stack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package orders

import (
	"fmt"
	"sync"
)

// Map represents an associative array or map abstract data type.
type Map struct {
	// mu Mutex protects data structures below.
	mu sync.Mutex

	// keys is the Set list of keys.
	keys []interface{}

	// store is the Set underlying store of values.
	store map[interface{}]interface{}
}

// NewOrderedMap creates a new empty Map.
func NewOrderedMap() OrderMap {
	m := &Map{
		keys:  make([]interface{}, 0),
		store: make(map[interface{}]interface{}),
	}

	return m
}

// Put adds items to the map.
//
// If a key is found in the map it replaces it value.
func (m *Map) Put(key interface{}, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.store[key]; !ok {
		m.keys = append(m.keys, key)
	}

	m.store[key] = value
}

// Get returns the value of a key from the Map.
func (m *Map) Get(key interface{}) (value interface{}, found bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	value, found = m.store[key]
	return value, found
}

// Remove deletes a key-value pair from the Map.
//
// If a key is not found in the map it doesn't fails, just does nothing.
func (m *Map) Remove(key interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check key exists
	if _, found := m.store[key]; !found {
		return
	}

	// Remove the value from the store
	delete(m.store, key)

	// Remove the key
	for i := range m.keys {
		if m.keys[i] == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

// Size return the map number of key-value pairs.
func (m *Map) Size() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.store)
}

// Empty return if the map in empty or not.
func (m *Map) Empty() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.store) == 0
}

// Keys return the keys in the map in insertion order.
func (m *Map) Keys() []interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.keys
}

// Values return the values in the map in insertion order.
func (m *Map) Values() []interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	values := make([]interface{}, len(m.store))
	for i, key := range m.keys {
		values[i] = m.store[key]
	}
	return values
}

// String implements Stringer interface.
//
// Prints the map string representation, a concatenated string of all its
// string representation values in insertion order.
func (m *Map) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []string
	for i, key := range m.keys {
		result = append(result, fmt.Sprintf("%d:%s", m.keys[i].(int), m.store[key]))
	}

	return fmt.Sprintf("%s", result)
}
