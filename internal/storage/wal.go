package storage

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "sync"
		"context"
)

// WALEntry represents a single operation in the log
type WALEntry struct {
		Index     int    `json:"index"`
    Operation string `json:"op"`    // "SET", "DELETE"
    Key       string `json:"key"`
    Value     string `json:"value"` // Empty for DELETE
}

// WAL handles write-ahead logging
type WAL struct {
    mu        sync.Mutex
    walFile   *os.File
    writer    *bufio.Writer
    encoder   *json.Encoder
		lastIndex int
}

// NewWAL creates or opens a WAL
func NewWAL(walFilePath string) (*WAL, error) {
    // Open file in append mode, create if doesn't exist
    file, err := os.OpenFile(walFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return nil, fmt.Errorf("failed to open WAL: %w", err)
    }

    return &WAL{
        walFile:    file,
        writer:  bufio.NewWriter(file),
        encoder: json.NewEncoder(file),
    }, nil
}

// Append writes an entry to the WAL
func (w *WAL) Append(op, key, value string) error {
    w.mu.Lock()
    defer w.mu.Unlock()
		
		w.lastIndex++

    entry := WALEntry{
				Index:     w.lastIndex,   
        Operation: op,
        Key:       key,
        Value:     value,
    }

    if err := w.encoder.Encode(entry); err != nil {
        return fmt.Errorf("failed to encode WAL entry: %w", err)
    }

    // Flush to OS buffer
    if err := w.writer.Flush(); err != nil {
        return fmt.Errorf("failed to flush WAL: %w", err)
    }

    // Force to disk
    if err := w.walFile.Sync(); err != nil {
        return fmt.Errorf("failed to sync WAL to disk: %w", err)
    }

    return nil
}

func (w *WAL) Close() error {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    if err := w.writer.Flush(); err != nil {
        return err
    }
    return w.walFile.Close()
}

// Replay reads the WAL and applies operations to the store
func (w *WAL) Replay(store *MemoryStore) error {
    w.mu.Lock()
    defer w.mu.Unlock()

    scanner := bufio.NewScanner(w.walFile)
    // Increase buffer size for large entries
    scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

    var entry WALEntry
    for scanner.Scan() {
        if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
            return fmt.Errorf("failed to parse WAL entry: %w", err)
        }

        // Apply the operation to memory store
        switch entry.Operation {
        case "SET":
            store.Put(context.Background(), entry.Key, entry.Value)
        case "DELETE":
            store.Delete(context.Background(), entry.Key)
        default:
            return fmt.Errorf("unknown operation in WAL: %s", entry.Operation)
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error reading WAL: %w", err)
    }

    return nil
}
