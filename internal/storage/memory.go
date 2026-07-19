package storage

import (
	"context"
	"maps"
	"sync"
)

type MemoryStore struct {
    mu    sync.RWMutex
    store map[string]string
}

func NewMemoryStore() *MemoryStore {
    return &MemoryStore{
        store: make(map[string]string),
    }
}

func (m *MemoryStore) Get(ctx context.Context, key string) (string, bool, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    val, ok := m.store[key]
    return val, ok, nil
}

func (m *MemoryStore) Put(ctx context.Context, key, value string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.store[key] = value
    return nil
}

func (m *MemoryStore) Delete(ctx context.Context, key string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    delete(m.store, key)
    return nil
}

func (m *MemoryStore) GetAll(ctx context.Context) (map[string]string, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    // Return a copy to avoid race conditions
    result := make(map[string]string, len(m.store))
		maps.Copy(result, m.store)

    return result, nil
}

func (m *MemoryStore) Close() error {
    return nil
}
