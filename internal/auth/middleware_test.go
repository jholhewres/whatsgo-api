package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jholhewres/whatsgo-api/internal/cache"
	"github.com/jholhewres/whatsgo-api/internal/store"
)

// --- mock store (only auth-relevant methods implemented) ---

type mockStore struct {
	store.Store // embed interface; panics on unimplemented calls
	instances   map[string]*store.Instance
	byToken     map[string]*store.Instance
	errOnLookup error
}

func newMockStore() *mockStore {
	return &mockStore{
		instances: make(map[string]*store.Instance),
		byToken:   make(map[string]*store.Instance),
	}
}

func (m *mockStore) addInstance(inst *store.Instance) {
	m.instances[inst.Name] = inst
	m.byToken[inst.Token] = inst
}

func (m *mockStore) GetInstanceByToken(_ context.Context, token string) (*store.Instance, error) {
	if m.errOnLookup != nil {
		return nil, m.errOnLookup
	}
	return m.byToken[token], nil
}

func (m *mockStore) GetInstanceByName(_ context.Context, name string) (*store.Instance, error) {
	if m.errOnLookup != nil {
		return nil, m.errOnLookup
	}
	return m.instances[name], nil
}

// --- helpers ---

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func doRequest(handler http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// --- RequireAuth tests ---

func TestRequireAuth_NoToken(t *testing.T) {
	ms := newMockStore()
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)

	handler := mw.RequireAuth(okHandler())
	rec := doRequest(handler, "GET", "/test", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_GlobalAPIKey(t *testing.T) {
	ms := newMockStore()
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)

	var gotGlobal bool
	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotGlobal = IsGlobalAuth(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rec := doRequest(handler, "GET", "/test", map[string]string{
		"X-API-Key": "test-global-key-1234",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !gotGlobal {
		t.Fatal("expected global auth context to be set")
	}
}

func TestRequireAuth_BearerToken(t *testing.T) {
	ms := newMockStore()
	ms.addInstance(&store.Instance{
		ID:    "inst-1",
		Name:  "my-instance",
		Token: "instance-token-abc",
	})
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)

	var gotInst *store.Instance
	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotInst = GetInstance(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rec := doRequest(handler, "GET", "/test", map[string]string{
		"Authorization": "Bearer instance-token-abc",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotInst == nil || gotInst.Name != "my-instance" {
		t.Fatal("expected instance context to be set with 'my-instance'")
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	ms := newMockStore()
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)

	handler := mw.RequireAuth(okHandler())
	rec := doRequest(handler, "GET", "/test", map[string]string{
		"X-API-Key": "wrong-token",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_StoreLookupError(t *testing.T) {
	ms := newMockStore()
	ms.errOnLookup = errors.New("db connection failed")
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)

	handler := mw.RequireAuth(okHandler())
	rec := doRequest(handler, "GET", "/test", map[string]string{
		"X-API-Key": "some-instance-token",
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestRequireAuth_CachesInstanceToken(t *testing.T) {
	ms := newMockStore()
	ms.addInstance(&store.Instance{
		ID:    "inst-1",
		Name:  "cached",
		Token: "cached-token",
	})
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)
	handler := mw.RequireAuth(okHandler())

	// First request populates cache
	doRequest(handler, "GET", "/test", map[string]string{
		"X-API-Key": "cached-token",
	})

	// Remove from store — cache should still work
	delete(ms.byToken, "cached-token")

	rec := doRequest(handler, "GET", "/test", map[string]string{
		"X-API-Key": "cached-token",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected cached auth to succeed, got %d", rec.Code)
	}
}

// --- RequireInstance tests ---

func TestRequireInstance_GlobalAuth(t *testing.T) {
	ms := newMockStore()
	ms.addInstance(&store.Instance{
		ID:   "inst-1",
		Name: "target",
	})
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)

	mux := http.NewServeMux()
	mux.Handle("GET /api/{name}/status", mw.RequireInstance(okHandler()))

	req := httptest.NewRequest("GET", "/api/target/status", nil)
	req.Header.Set("X-API-Key", "test-global-key-1234")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireInstance_GlobalAuth_NotFound(t *testing.T) {
	ms := newMockStore()
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)

	mux := http.NewServeMux()
	mux.Handle("GET /api/{name}/status", mw.RequireInstance(okHandler()))

	req := httptest.NewRequest("GET", "/api/nonexistent/status", nil)
	req.Header.Set("X-API-Key", "test-global-key-1234")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRequireInstance_InstanceToken_MatchingName(t *testing.T) {
	ms := newMockStore()
	ms.addInstance(&store.Instance{
		ID:    "inst-1",
		Name:  "my-inst",
		Token: "my-inst-token",
	})
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)

	mux := http.NewServeMux()
	mux.Handle("GET /api/{name}/status", mw.RequireInstance(okHandler()))

	req := httptest.NewRequest("GET", "/api/my-inst/status", nil)
	req.Header.Set("X-API-Key", "my-inst-token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireInstance_InstanceToken_WrongName(t *testing.T) {
	ms := newMockStore()
	ms.addInstance(&store.Instance{
		ID:    "inst-1",
		Name:  "my-inst",
		Token: "my-inst-token",
	})
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)

	mux := http.NewServeMux()
	mux.Handle("GET /api/{name}/status", mw.RequireInstance(okHandler()))

	req := httptest.NewRequest("GET", "/api/other-inst/status", nil)
	req.Header.Set("X-API-Key", "my-inst-token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// --- Cache invalidation ---

func TestInvalidateToken_RemovesFromCache(t *testing.T) {
	ms := newMockStore()
	ms.addInstance(&store.Instance{
		ID:    "inst-1",
		Name:  "rm-me",
		Token: "rm-token",
	})
	mw := NewMiddleware(ms, "test-global-key-1234", discardLogger)
	handler := mw.RequireAuth(okHandler())

	// Populate cache
	doRequest(handler, "GET", "/test", map[string]string{
		"X-API-Key": "rm-token",
	})

	// Invalidate + remove from store
	mw.InvalidateToken("rm-token")
	delete(ms.byToken, "rm-token")

	rec := doRequest(handler, "GET", "/test", map[string]string{
		"X-API-Key": "rm-token",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after invalidation, got %d", rec.Code)
	}
}

func TestCleanupCache_RemovesExpired(t *testing.T) {
	ms := newMockStore()
	ms.addInstance(&store.Instance{
		ID:    "inst-1",
		Name:  "expiry-test",
		Token: "expiry-token",
	})

	// Create middleware with a very short cache TTL for testing
	mw := &Middleware{
		store:         ms,
		globalKey:     "test-global-key-1234",
		logger:        discardLogger,
		instanceCache: newShortTTLCache(),
	}
	handler := mw.RequireAuth(okHandler())

	// Populate cache
	doRequest(handler, "GET", "/test", map[string]string{
		"X-API-Key": "expiry-token",
	})

	// Wait for TTL to expire + cleanup
	time.Sleep(150 * time.Millisecond)
	mw.CleanupCache()

	// Remove from store — cache should be expired
	delete(ms.byToken, "expiry-token")

	rec := doRequest(handler, "GET", "/test", map[string]string{
		"X-API-Key": "expiry-token",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after cache cleanup, got %d", rec.Code)
	}
}

// --- helper for short TTL cache ---

func newShortTTLCache() *cache.Cache[string, *store.Instance] {
	return cache.New[string, *store.Instance](100 * time.Millisecond)
}

// --- extractToken tests ---

func TestExtractToken_XAPIKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "my-key")
	if got := extractToken(req); got != "my-key" {
		t.Fatalf("expected 'my-key', got '%s'", got)
	}
}

func TestExtractToken_Bearer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer my-bearer")
	if got := extractToken(req); got != "my-bearer" {
		t.Fatalf("expected 'my-bearer', got '%s'", got)
	}
}

func TestExtractToken_AuthorizationNonBearer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic abc123")
	if got := extractToken(req); got != "Basic abc123" {
		t.Fatalf("expected 'Basic abc123', got '%s'", got)
	}
}

func TestExtractToken_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if got := extractToken(req); got != "" {
		t.Fatalf("expected empty, got '%s'", got)
	}
}

func TestExtractToken_AuthorizationTakesPrecedence(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer auth-token")
	req.Header.Set("X-API-Key", "api-key-token")
	if got := extractToken(req); got != "auth-token" {
		t.Fatalf("expected Authorization header to take precedence, got '%s'", got)
	}
}
