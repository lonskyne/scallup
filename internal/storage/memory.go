package storage

import (
	"context"
	"maps"
	"sync"
	"fmt"
)

type MemoryStore struct {
    mu    sync.RWMutex
    store map[string]string
		wal   *WAL
}

func NewMemoryStore(walFilePath string) (*MemoryStore, error) {
	 	wal, err := NewWAL(walFilePath)
    if err != nil {
			return nil, fmt.Errorf("failed to create WAL object: %w", err)
    }

    return &MemoryStore{
        store: make(map[string]string),
				wal:   wal,
    }, nil
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
    
		if err := m.wal.Append("SET", key, value); err != nil {
        return fmt.Errorf("failed to write to WAL: %w", err)
    }

    m.store[key] = value
    return nil
}

func (m *MemoryStore) Delete(ctx context.Context, key string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
		if err := m.wal.Append("DELETE", key, ""); err != nil {
        return fmt.Errorf("failed to write to WAL: %w", err)
    }

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
    return m.wal.Close()
}
