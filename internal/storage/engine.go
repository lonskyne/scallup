package storage

import "context"

// Engine defines the interface for our key-value storage
type Engine interface {
    Get(ctx context.Context, key string) (string, bool, error)
    Put(ctx context.Context, key, value string) error
    Delete(ctx context.Context, key string) error
    GetAll(ctx context.Context) (map[string]string, error)
    Close() error
}
