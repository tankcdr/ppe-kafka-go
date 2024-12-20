package main

import "sync"

type SimpleDatabase struct {
	mu   sync.RWMutex
	data map[string]struct{}
}

func NewSimpleDatabase() *SimpleDatabase {
	return &SimpleDatabase{
		data: make(map[string]struct{}),
	}
}

func (db *SimpleDatabase) Add(value string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data[value] = struct{}{}
}

func (db *SimpleDatabase) Exists(value string) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.data[value]
	return exists
}
