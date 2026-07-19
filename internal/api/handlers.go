package api

import (
    "encoding/json"
    "net/http"
    "strings"

		"github.com/lonskyne/scallup/internal/storage"
		"github.com/lonskyne/scallup/pkg/types"
)

type Handlers struct {
    storage storage.Engine
}

func NewHandlers(storage storage.Engine) *Handlers {
    return &Handlers{storage: storage}
}

func (h *Handlers) GetKey(w http.ResponseWriter, r *http.Request) {
    // Extract key from URL path
    path := strings.TrimPrefix(r.URL.Path, "/api/v1/key/")
    if path == "" {
        http.Error(w, "Key is required", http.StatusBadRequest)
        return
    }

    value, ok, err := h.storage.Get(r.Context(), path)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !ok {
        http.Error(w, "Key not found", http.StatusNotFound)
        return
    }

    response := types.KeyValue{
        Key:   path,
        Value: value,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *Handlers) PutKey(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/api/v1/key/")
    if path == "" {
        http.Error(w, "Key is required", http.StatusBadRequest)
        return
    }

    var kv types.KeyValue
    if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.storage.Put(r.Context(), path, kv.Value); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

func (h *Handlers) DeleteKey(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/api/v1/key/")
    if path == "" {
        http.Error(w, "Key is required", http.StatusBadRequest)
        return
    }

    if err := h.storage.Delete(r.Context(), path); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetAllKeys(w http.ResponseWriter, r *http.Request) {
    keys, err := h.storage.GetAll(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(keys)
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "OK",
    })
}
