package auth

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jholhewres/whatsgo-api/internal/cache"
	"github.com/jholhewres/whatsgo-api/internal/store"
)

type Middleware struct {
	store         store.Store
	globalKey     string
	logger        *slog.Logger
	instanceCache *cache.Cache[string, *store.Instance]
}

func NewMiddleware(s store.Store, globalKey string, logger *slog.Logger) *Middleware {
	return &Middleware{
		store:         s,
		globalKey:     globalKey,
		logger:        logger,
		instanceCache: cache.New[string, *store.Instance](2 * time.Minute),
	}
}

func (m *Middleware) CleanupCache() {
	m.instanceCache.Cleanup()
}

// InvalidateToken removes a token from the cache (e.g., on delete/logout).
func (m *Middleware) InvalidateToken(token string) {
	m.instanceCache.Invalidate(token)
}

// RequireAuth validates the API key from the request.
// If it matches the global key, grants full access.
// If it matches an instance token, loads that instance into context.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, `{"error":"missing authorization"}`, http.StatusUnauthorized)
			return
		}

		// Check global API key
		if m.globalKey != "" && token == m.globalKey {
			ctx := SetGlobalAuth(r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Check instance token
		inst, ok := m.instanceCache.Get(token)
		if !ok {
			var err error
			inst, err = m.store.GetInstanceByToken(r.Context(), token)
			if err != nil {
				m.logger.Error("auth lookup failed", "error", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				return
			}
			if inst != nil {
				m.instanceCache.Set(token, inst)
			}
		}

		if inst == nil {
			http.Error(w, `{"error":"invalid api key"}`, http.StatusUnauthorized)
			return
		}

		ctx := SetInstance(r.Context(), inst)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireInstance wraps RequireAuth and verifies the auth token
// has access to the instance specified by {name} in the URL path.
func (m *Middleware) RequireInstance(next http.Handler) http.Handler {
	return m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.Error(w, `{"error":"instance name required"}`, http.StatusBadRequest)
			return
		}

		// Global auth can access any instance
		if IsGlobalAuth(r.Context()) {
			inst, err := m.store.GetInstanceByName(r.Context(), name)
			if err != nil {
				m.logger.Error("instance lookup failed", "error", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				return
			}
			if inst == nil {
				http.Error(w, `{"error":"instance not found"}`, http.StatusNotFound)
				return
			}
			ctx := SetInstance(r.Context(), inst)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Instance-scoped auth: verify the token belongs to this instance
		inst := GetInstance(r.Context())
		if inst == nil || inst.Name != name {
			http.Error(w, `{"error":"access denied to this instance"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}))
}

func extractToken(r *http.Request) string {
	// Check Authorization: Bearer <token>
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		return auth
	}

	// Check X-API-Key header
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}

	return ""
}
