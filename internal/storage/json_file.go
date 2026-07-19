package storage

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"sync"
	"errors"

	"encoding/json"
)

type JSONFileStore struct {
    mu       sync.RWMutex
    store    map[string]string
		file     *os.File
		encoder  *json.Encoder
		wal      *WAL
}

func NewJSONFileStore(filePath string, walFilePath string) (*JSONFileStore, error) {
	wal, err := NewWAL(walFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create WAL object: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}

	encoder := json.NewEncoder(file)

	store := &JSONFileStore{
        store:   make(map[string]string),
				file:    file,
				encoder: encoder,
				wal:     wal,
  	}

	decoder := json.NewDecoder(file)

  store.mu.Lock()
  defer store.mu.Unlock()

	err = store.loadJSONFileIfExists(decoder);
	if err != nil {
		return nil, fmt.Errorf("Failed to load JSON file: %w", err)
	}

	return store, nil
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

		if err := m.wal.Append("SET", key, value); err != nil {
        return fmt.Errorf("failed to write to WAL: %w", err)
    }
    
    m.store[key] = value
    return m.writeJSONFile()
}

func (m *JSONFileStore) Delete(ctx context.Context, key string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
		if err := m.wal.Append("DELETE", key, ""); err != nil {
        return fmt.Errorf("failed to write to WAL: %w", err)
    }

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
		err := m.file.Close()
		if err != nil {
			return err
		}

    return m.wal.Close()
}

func (m *JSONFileStore) writeJSONFile() error {
	if err := m.file.Truncate(0); err != nil {
			return err
	}
	
	if _, err := m.file.Seek(0, 0); err != nil {
			return err
	}

	if err := m.encoder.Encode(m.store); err != nil {
		return err
	}

	return m.file.Sync()
}

func (m *JSONFileStore) loadJSONFileIfExists(decoder *json.Decoder) error {
	if _, err := m.file.Seek(0, 0); err != nil {
    return err
  }

	if err := decoder.Decode(&m.store); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("failed to load JSON file store: %w", err)
	}

	return nil
}
