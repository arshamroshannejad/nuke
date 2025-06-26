package nuke

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter()
	if r == nil {
		t.Error("NewRouter() returned nil")
	}
	if r.mux == nil {
		t.Error("NewRouter() didn't initialize ServeMux")
	}
	if len(r.globalChain) != 0 {
		t.Error("NewRouter() should start with empty middleware chain")
	}
}

func TestRouter_Use(t *testing.T) {
	t.Run("global middleware", func(t *testing.T) {
		r := NewRouter()
		mwCalled := false
		mw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mwCalled = true
				next.ServeHTTP(w, r)
			})
		}
		r.Use(mw)
		if len(r.globalChain) != 1 {
			t.Error("Use() didn't add middleware to global chain")
		}
		r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if !mwCalled {
			t.Error("Global middleware was not called")
		}
	})

	t.Run("group middleware", func(t *testing.T) {
		r := NewRouter()
		groupMwCalled := false
		groupMw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				groupMwCalled = true
				next.ServeHTTP(w, r)
			})
		}
		r.Group(func(sub *Router) {
			sub.Use(groupMw)
			sub.HandleFunc("/group", func(w http.ResponseWriter, r *http.Request) {})
		})
		req := httptest.NewRequest("GET", "/group", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if !groupMwCalled {
			t.Error("Group middleware was not called for group route")
		}
		groupMwCalled = false
		r.HandleFunc("/normal", func(w http.ResponseWriter, r *http.Request) {})
		req = httptest.NewRequest("GET", "/normal", nil)
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if groupMwCalled {
			t.Error("Group middleware was called for non-group route")
		}
	})
}

func TestRouter_Group(t *testing.T) {
	r := NewRouter()
	called := false
	r.Group(func(sub *Router) {
		if !sub.isSubRouter {
			t.Error("Group router should be marked as sub-router")
		}
		if sub.mux != r.mux {
			t.Error("Group router should share the same mux with parent")
		}
		called = true
	})
	if !called {
		t.Error("Group callback was not called")
	}
}

func TestRouter_HandleFunc(t *testing.T) {
	r := NewRouter()
	handlerCalled := false
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if !handlerCalled {
		t.Error("Handler function was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRouter_Handle(t *testing.T) {
	r := NewRouter()
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusCreated)
	})
	r.Handle("/test", handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if !handlerCalled {
		t.Error("Handler was not called")
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestRouter_ServeHTTP(t *testing.T) {
	t.Run("global middleware order", func(t *testing.T) {
		r := NewRouter()
		var order []string
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "first")
				next.ServeHTTP(w, r)
			})
		})
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "second")
				next.ServeHTTP(w, r)
			})
		})
		r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
		})
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		expected := []string{"first", "second", "handler"}
		if len(order) != len(expected) {
			t.Fatalf("Expected %d calls, got %d", len(expected), len(order))
		}
		for i, v := range expected {
			if order[i] != v {
				t.Errorf("At position %d, expected %s, got %s", i, v, order[i])
			}
		}
	})
	t.Run("group middleware order", func(t *testing.T) {
		r := NewRouter()
		var order []string
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "global")
				next.ServeHTTP(w, r)
			})
		})
		r.Group(func(sub *Router) {
			sub.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, "group")
					next.ServeHTTP(w, r)
				})
			})
			sub.HandleFunc("/group", func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "group-handler")
			})
		})
		req := httptest.NewRequest("GET", "/group", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		expected := []string{"global", "group", "group-handler"}
		if len(order) != len(expected) {
			t.Fatalf("Expected %d calls, got %d", len(expected), len(order))
		}
		for i, v := range expected {
			if order[i] != v {
				t.Errorf("At position %d, expected %s, got %s", i, v, order[i])
			}
		}
	})
	t.Run("not found", func(t *testing.T) {
		r := NewRouter()
		req := httptest.NewRequest("GET", "/not-found", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}

func BenchmarkNewRouter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewRouter()
	}
}

func BenchmarkRouter_Use(b *testing.B) {
	r := NewRouter()
	mw := func(next http.Handler) http.Handler {
		return next
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Use(mw)
	}
}

func BenchmarkRouter_Group(b *testing.B) {
	r := NewRouter()
	fn := func(sub *Router) {}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Group(fn)
	}
}
func BenchmarkRouter_HandleFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := NewRouter()
		handler := func(w http.ResponseWriter, r *http.Request) {}
		r.HandleFunc("/test", handler)
	}
}

func BenchmarkRouter_Handle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := NewRouter()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		r.Handle("/test", handler)
	}
}

func BenchmarkRouter_ServeHTTP(b *testing.B) {
	r := NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rec, req)
	}
}
