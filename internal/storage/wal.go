package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// WALEntry represents a single operation in the log
type WALEntry struct {
	Index     int    `json:"index"`
	Operation string `json:"op"` // "SET", "DELETE"
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
	file, err := os.OpenFile(walFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL: %w", err)
	}

	writer := bufio.NewWriter(file)

	prevLastIndex, err := getLastIndexFromWAL(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get last index from WAL: %w", err)
	}

	return &WAL{
		walFile:   file,
		writer:    writer,
		encoder:   json.NewEncoder(writer),
		lastIndex: prevLastIndex,
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

func getLastIndexFromWAL(file *os.File) (int, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return -1, fmt.Errorf("failed to get WAL stat: %w", err)
	}

	if fileInfo.Size() <= 0 {
		return 0, nil
	}

	prevLastIndex := 0
	readOffset := fileInfo.Size() - 1
	var buf [1]byte

	for buf[0] != '{' && readOffset >= 0 {
		_, err := file.ReadAt(buf[:], readOffset)
		if err != nil {
			return -1, fmt.Errorf("failed to read WAL lastIndex: %w", err)
		}

		readOffset--
	}

	start := readOffset + 1
	length := fileInfo.Size() - start
	lastLine := make([]byte, length)

	_, err = file.ReadAt(lastLine, start)
	if err != nil {
		return -1, fmt.Errorf("failed to read WAL lastLine: %w", err)
	}

	var lastEntry WALEntry

	err = json.Unmarshal(lastLine, &lastEntry)
	if err != nil {
		return -1, fmt.Errorf("failed to unmarshal last entry in WAL: %w", err)
	}

	prevLastIndex = lastEntry.Index

	return prevLastIndex, nil
}
