package db

import (
	"sync"
)

// SimpleInMemoryDatabase is a thread-safe in-memory key-value store.
// K must be comparable (e.g., string, int), and V can be any type.
type SimpleInMemoryDatabase[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

// NewSimpleInMemoryDatabase creates a new generic in-memory database.
func NewSimpleInMemoryDatabase[K comparable, V any]() *SimpleInMemoryDatabase[K, V] {
	return &SimpleInMemoryDatabase[K, V]{
		data: make(map[K]V),
	}
}

// Add inserts or updates a value for the given key.
func (db *SimpleInMemoryDatabase[K, V]) Add(key K, value V) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data[key] = value
}

// Exists checks if a key is in the database.
func (db *SimpleInMemoryDatabase[K, V]) Exists(key K) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.data[key]
	return exists
}

// Get retrieves the value associated with the given key.
// It returns the value and a boolean indicating if the key exists.
func (db *SimpleInMemoryDatabase[K, V]) Get(key K) (V, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	val, exists := db.data[key]
	return val, exists
}

// Delete removes a key (and its value) from the database.
func (db *SimpleInMemoryDatabase[K, V]) Delete(key K) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.data, key)
}
