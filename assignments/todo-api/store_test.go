package main

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestMemoryStore_CreateGet(t *testing.T) {
	for _, tc := range []struct {
		name  string
		title string
		done  bool
	}{
		{"not done", "write tests", false},
		{"done", "ship it", true},
		{"empty title is stored as given", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewMemoryStore()
			ctx := context.Background()

			created, err := s.Create(ctx, tc.title, tc.done)
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			if created.ID != 1 {
				t.Errorf("first ID = %d, want 1", created.ID)
			}
			got, err := s.Get(ctx, created.ID)
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			if got != (Todo{ID: 1, Title: tc.title, Done: tc.done}) {
				t.Errorf("Get = %+v", got)
			}
		})
	}
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   int
	}{
		{"missing", 99},
		{"zero", 0},
		{"negative", -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewMemoryStore()
			if _, err := s.Create(context.Background(), "a", false); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Get(context.Background(), tc.id); !errors.Is(err, ErrNotFound) {
				t.Errorf("Get err = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("removes only target and keeps order", func(t *testing.T) {
		s := NewMemoryStore()
		for _, title := range []string{"a", "b", "c"} {
			if _, err := s.Create(ctx, title, false); err != nil {
				t.Fatal(err)
			}
		}
		if err := s.Delete(ctx, 2); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := s.Get(ctx, 2); !errors.Is(err, ErrNotFound) {
			t.Errorf("Get after delete err = %v, want ErrNotFound", err)
		}
		list, err := s.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != 2 || list[0].ID != 1 || list[1].ID != 3 {
			t.Errorf("List = %+v, want IDs [1 3]", list)
		}
	})

	t.Run("ids are not reused", func(t *testing.T) {
		s := NewMemoryStore()
		first, _ := s.Create(ctx, "a", false)
		if err := s.Delete(ctx, first.ID); err != nil {
			t.Fatal(err)
		}
		second, _ := s.Create(ctx, "b", false)
		if second.ID == first.ID {
			t.Errorf("ID %d reused", second.ID)
		}
	})

	for _, tc := range []struct {
		name string
		id   int
	}{
		{"missing", 99},
		{"zero", 0},
		{"negative", -5},
	} {
		t.Run("not found "+tc.name, func(t *testing.T) {
			s := NewMemoryStore()
			if err := s.Delete(ctx, tc.id); !errors.Is(err, ErrNotFound) {
				t.Errorf("Delete err = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestMemoryStore_List(t *testing.T) {
	ctx := context.Background()

	t.Run("empty is non-nil", func(t *testing.T) {
		list, err := NewMemoryStore().List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if list == nil {
			t.Error("List returned nil slice, want empty non-nil")
		}
		if len(list) != 0 {
			t.Errorf("len = %d, want 0", len(list))
		}
	})

	t.Run("insertion order", func(t *testing.T) {
		s := NewMemoryStore()
		want := []string{"c", "a", "b"}
		for _, title := range want {
			if _, err := s.Create(ctx, title, false); err != nil {
				t.Fatal(err)
			}
		}
		list, _ := s.List(ctx)
		for i, title := range want {
			if list[i].Title != title {
				t.Errorf("list[%d] = %q, want %q", i, list[i].Title, title)
			}
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		s := NewMemoryStore()
		if _, err := s.Create(ctx, "orig", false); err != nil {
			t.Fatal(err)
		}
		list, _ := s.List(ctx)
		list[0].Title = "mutated"
		_ = append(list[:0], Todo{ID: 42})

		got, err := s.Get(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.Title != "orig" {
			t.Errorf("store mutated via List result: %+v", got)
		}
	})
}

func TestMemoryStore_CancelledContext(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.Create(context.Background(), "a", false); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, tc := range []struct {
		name string
		call func() error
	}{
		{"Create", func() error { _, err := s.Create(ctx, "b", false); return err }},
		{"Get", func() error { _, err := s.Get(ctx, 1); return err }},
		{"List", func() error { _, err := s.List(ctx); return err }},
		{"Delete", func() error { return s.Delete(ctx, 1) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, context.Canceled) {
				t.Errorf("err = %v, want context.Canceled", err)
			}
		})
	}

	t.Run("cancelled create does not store", func(t *testing.T) {
		list, _ := s.List(context.Background())
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
	})
}

func TestMemoryStore_ConcurrentCreateUniqueIDs(t *testing.T) {
	const n = 100
	s := NewMemoryStore()
	ids := make([]int, n)
	errs := make([]error, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			todo, err := s.Create(context.Background(), "x", false)
			ids[i], errs[i] = todo.ID, err
		}(i)
	}
	wg.Wait()

	seen := make(map[int]bool, n)
	for i, id := range ids {
		if errs[i] != nil {
			t.Fatalf("Create %d: %v", i, errs[i])
		}
		if seen[id] {
			t.Fatalf("duplicate ID %d", id)
		}
		seen[id] = true
	}
	list, _ := s.List(context.Background())
	if len(list) != n {
		t.Errorf("len = %d, want %d", len(list), n)
	}
}
