package main

import (
	"context"
	"errors"
	"sync"
)

// ErrNotFound is returned when no todo has the requested ID.
var ErrNotFound = errors.New("todo not found")

// Todo is a single task.
type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Store is the persistence boundary. Handlers depend on this, not on MemoryStore.
type Store interface {
	Create(ctx context.Context, title string, done bool) (Todo, error)
	Get(ctx context.Context, id int) (Todo, error)
	List(ctx context.Context) ([]Todo, error)
	Delete(ctx context.Context, id int) error
}

// MemoryStore keeps todos in insertion order. Safe for concurrent use.
type MemoryStore struct {
	mu     sync.Mutex
	nextID int
	todos  []Todo
}

// NewMemoryStore returns an empty store whose first ID is 1.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{nextID: 1}
}

func (s *MemoryStore) Create(ctx context.Context, title string, done bool) (Todo, error) {
	if err := ctx.Err(); err != nil {
		return Todo{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Todo{ID: s.nextID, Title: title, Done: done}
	s.nextID++
	s.todos = append(s.todos, t)
	return t, nil
}

func (s *MemoryStore) Get(ctx context.Context, id int) (Todo, error) {
	if err := ctx.Err(); err != nil {
		return Todo{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if i := s.indexOf(id); i >= 0 {
		return s.todos[i], nil
	}
	return Todo{}, ErrNotFound
}

func (s *MemoryStore) List(ctx context.Context) ([]Todo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Todo, len(s.todos))
	copy(out, s.todos)
	return out, nil
}

func (s *MemoryStore) Delete(ctx context.Context, id int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(id)
	if i < 0 {
		return ErrNotFound
	}
	s.todos = append(s.todos[:i], s.todos[i+1:]...)
	return nil
}

// indexOf returns the position of id in s.todos, or -1. Caller must hold s.mu.
// Ids 0 and below never exist, so they fall through to -1 (404 at the handler).
func (s *MemoryStore) indexOf(id int) int {
	for i, t := range s.todos {
		if t.ID == id {
			return i
		}
	}
	return -1
}
