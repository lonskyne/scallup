package storage

import (
	"context"
	"maps"
	"os"
	"sync"

	"encoding/json"
)

type JSONFileStore struct {
    mu       sync.RWMutex
    store    map[string]string
		filePath string
}

func NewJSONFileStore(filePath string) *JSONFileStore {
	store := &JSONFileStore{
        store:    make(map[string]string),
				filePath: filePath,
    }

	store.loadJSONFileIfExists();

	return store
}

func (m *JSONFileStore) Get(ctx context.Context, key string) (string, bool, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    val, ok := m.store[key]
    return val, ok, nil
}

func (m *JSONFileStore) Put(ctx context.Context, key, value string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.store[key] = value
    return m.writeJSONFile()
}

func (m *JSONFileStore) Delete(ctx context.Context, key string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    delete(m.store, key)
    return m.writeJSONFile()
}

func (m *JSONFileStore) GetAll(ctx context.Context) (map[string]string, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    // Return a copy to avoid race conditions
    result := make(map[string]string, len(m.store))
		maps.Copy(result, m.store)

    return result, nil
}

func (m *JSONFileStore) Close() error {
    return nil
}

func (m *JSONFileStore) writeJSONFile() error {
	file, err := os.Create(m.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(m.store); err != nil {
		return err
	}

	return nil
}

func (m *JSONFileStore) loadJSONFileIfExists() error {
	file, err := os.Open(m.filePath)
	if err != nil{
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&m.store); err != nil {
		return err
	}

	return nil
}
