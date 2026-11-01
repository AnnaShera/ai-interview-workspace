package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestMux returns a router over a fresh store.
func newTestMux() *http.ServeMux {
	return newMux(NewHandlers(NewMemoryStore()))
}

func do(t *testing.T, mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

func assertJSONError(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("error body not JSON: %q", rec.Body.String())
	}
	if msg, ok := m["error"].(string); !ok || msg == "" {
		t.Errorf("body %v lacks non-empty \"error\" string", m)
	}
}

func TestCreate(t *testing.T) {
	for _, tc := range []struct {
		name       string
		body       string
		wantStatus int
		wantTitle  string
		wantDone   bool
	}{
		{"valid", `{"title":"buy milk"}`, http.StatusCreated, "buy milk", false},
		{"done accepted", `{"title":"x","done":true}`, http.StatusCreated, "x", true},
		{"title trimmed", `{"title":"  padded  "}`, http.StatusCreated, "padded", false},
		{"100 multibyte runes ok", `{"title":"` + strings.Repeat("é", 100) + `"}`, http.StatusCreated, strings.Repeat("é", 100), false},
		{"101 runes rejected", `{"title":"` + strings.Repeat("é", 101) + `"}`, http.StatusBadRequest, "", false},
		{"whitespace-only title", `{"title":"   "}`, http.StatusBadRequest, "", false},
		{"empty title", `{"title":""}`, http.StatusBadRequest, "", false},
		{"missing title", `{"done":true}`, http.StatusBadRequest, "", false},
		{"unknown field", `{"title":"x","extra":1}`, http.StatusBadRequest, "", false},
		{"bad JSON", `{"title":`, http.StatusBadRequest, "", false},
		{"empty body", ``, http.StatusBadRequest, "", false},
		{"wrong title type", `{"title":5}`, http.StatusBadRequest, "", false},
		{"trailing data", `{"title":"x"}{"title":"y"}`, http.StatusBadRequest, "", false},
		{"trailing garbage", `{"title":"x"} junk`, http.StatusBadRequest, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mux := newTestMux()
			rec := do(t, mux, http.MethodPost, "/todos", tc.body)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, tc.wantStatus, rec.Body)
			}
			if tc.wantStatus != http.StatusCreated {
				assertJSONError(t, rec)
				if list := decode[[]Todo](t, do(t, mux, http.MethodGet, "/todos", "")); len(list) != 0 {
					t.Errorf("rejected request still stored %d todos", len(list))
				}
				return
			}
			got := decode[Todo](t, rec)
			if got.ID != 1 || got.Title != tc.wantTitle || got.Done != tc.wantDone {
				t.Errorf("created = %+v", got)
			}
		})
	}
}

func TestGet(t *testing.T) {
	for _, tc := range []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"existing", "/todos/1", http.StatusOK},
		{"unknown id", "/todos/999", http.StatusNotFound},
		{"zero id", "/todos/0", http.StatusNotFound},
		{"negative id", "/todos/-1", http.StatusNotFound},
		{"non-numeric id", "/todos/abc", http.StatusBadRequest},
		{"float id", "/todos/1.5", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mux := newTestMux()
			if rec := do(t, mux, http.MethodPost, "/todos", `{"title":"seed"}`); rec.Code != http.StatusCreated {
				t.Fatalf("seed status = %d", rec.Code)
			}
			rec := do(t, mux, http.MethodGet, tc.path, "")
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantStatus == http.StatusOK {
				if got := decode[Todo](t, rec); got.Title != "seed" {
					t.Errorf("got %+v", got)
				}
				return
			}
			assertJSONError(t, rec)
		})
	}
}

func TestList(t *testing.T) {
	t.Run("empty encodes as [] not null", func(t *testing.T) {
		rec := do(t, newTestMux(), http.MethodGet, "/todos", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d", rec.Code)
		}
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Errorf("body = %q, want []", got)
		}
	})

	t.Run("insertion order", func(t *testing.T) {
		mux := newTestMux()
		titles := []string{"first", "second", "third"}
		for _, title := range titles {
			do(t, mux, http.MethodPost, "/todos", `{"title":"`+title+`"}`)
		}
		rec := do(t, mux, http.MethodGet, "/todos", "")
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		list := decode[[]Todo](t, rec)
		if len(list) != len(titles) {
			t.Fatalf("len = %d, want %d", len(list), len(titles))
		}
		for i, title := range titles {
			if list[i].Title != title {
				t.Errorf("list[%d] = %q, want %q", i, list[i].Title, title)
			}
		}
	})
}

func TestDelete(t *testing.T) {
	for _, tc := range []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"existing", "/todos/1", http.StatusNoContent},
		{"unknown id", "/todos/999", http.StatusNotFound},
		{"zero id", "/todos/0", http.StatusNotFound},
		{"negative id", "/todos/-1", http.StatusNotFound},
		{"non-numeric id", "/todos/abc", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mux := newTestMux()
			do(t, mux, http.MethodPost, "/todos", `{"title":"seed"}`)

			rec := do(t, mux, http.MethodDelete, tc.path, "")
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantStatus == http.StatusNoContent {
				if rec.Body.Len() != 0 {
					t.Errorf("204 body = %q, want empty", rec.Body)
				}
				return
			}
			assertJSONError(t, rec)
		})
	}

	t.Run("delete then get is 404", func(t *testing.T) {
		mux := newTestMux()
		do(t, mux, http.MethodPost, "/todos", `{"title":"seed"}`)
		if rec := do(t, mux, http.MethodDelete, "/todos/1", ""); rec.Code != http.StatusNoContent {
			t.Fatalf("delete status = %d", rec.Code)
		}
		rec := do(t, mux, http.MethodGet, "/todos/1", "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("get status = %d, want 404", rec.Code)
		}
		assertJSONError(t, rec)
	})

	t.Run("second delete is 404", func(t *testing.T) {
		mux := newTestMux()
		do(t, mux, http.MethodPost, "/todos", `{"title":"seed"}`)
		do(t, mux, http.MethodDelete, "/todos/1", "")
		if rec := do(t, mux, http.MethodDelete, "/todos/1", ""); rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})
}
