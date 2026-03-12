package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jholhewres/whatsgo-api/internal/auth"
	"github.com/jholhewres/whatsgo-api/internal/config"
	"github.com/jholhewres/whatsgo-api/internal/store"
	"github.com/jholhewres/whatsgo-api/internal/whatsapp"
)

type Handlers struct {
	store   store.Store
	manager *whatsapp.Manager
	authMw  *auth.Middleware
	cfg     *config.Config
	logger  *slog.Logger
}

func NewHandlers(
	s store.Store,
	manager *whatsapp.Manager,
	authMw *auth.Middleware,
	cfg *config.Config,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		store:   s,
		manager: manager,
		authMw:  authMw,
		cfg:     cfg,
		logger:  logger,
	}
}

func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	requireAuth := h.authMw.RequireAuth
	requireInstance := h.authMw.RequireInstance

	// Instance management
	mux.Handle("POST /api/v1/instance/create", requireAuth(http.HandlerFunc(h.HandleCreateInstance)))
	mux.Handle("POST /api/v1/instance/{name}/connect", requireInstance(http.HandlerFunc(h.HandleConnect)))
	mux.Handle("POST /api/v1/instance/{name}/restart", requireInstance(http.HandlerFunc(h.HandleRestart)))
	mux.Handle("GET /api/v1/instance/{name}/status", requireInstance(http.HandlerFunc(h.HandleStatus)))
	mux.Handle("GET /api/v1/instance", requireAuth(http.HandlerFunc(h.HandleListInstances)))
	mux.Handle("DELETE /api/v1/instance/{name}/logout", requireInstance(http.HandlerFunc(h.HandleLogout)))
	mux.Handle("DELETE /api/v1/instance/{name}", requireInstance(http.HandlerFunc(h.HandleDeleteInstance)))

	// Messages
	mux.Handle("POST /api/v1/instance/{name}/message/send-text", requireInstance(http.HandlerFunc(h.HandleSendText)))
	mux.Handle("POST /api/v1/instance/{name}/message/send-media", requireInstance(http.HandlerFunc(h.HandleSendMedia)))
	mux.Handle("POST /api/v1/instance/{name}/message/send-location", requireInstance(http.HandlerFunc(h.HandleSendLocation)))
	mux.Handle("POST /api/v1/instance/{name}/message/send-contact", requireInstance(http.HandlerFunc(h.HandleSendContact)))
	mux.Handle("POST /api/v1/instance/{name}/message/send-reaction", requireInstance(http.HandlerFunc(h.HandleSendReaction)))
	mux.Handle("POST /api/v1/instance/{name}/message/send-sticker", requireInstance(http.HandlerFunc(h.HandleSendSticker)))

	// Chat
	mux.Handle("POST /api/v1/instance/{name}/chat/check-number", requireInstance(http.HandlerFunc(h.HandleCheckNumber)))
	mux.Handle("POST /api/v1/instance/{name}/chat/mark-read", requireInstance(http.HandlerFunc(h.HandleMarkRead)))
	mux.Handle("POST /api/v1/instance/{name}/chat/delete-message", requireInstance(http.HandlerFunc(h.HandleDeleteMessage)))
	mux.Handle("POST /api/v1/instance/{name}/chat/edit-message", requireInstance(http.HandlerFunc(h.HandleEditMessage)))
	mux.Handle("POST /api/v1/instance/{name}/chat/send-presence", requireInstance(http.HandlerFunc(h.HandleSendPresence)))
	mux.Handle("POST /api/v1/instance/{name}/chat/block", requireInstance(http.HandlerFunc(h.HandleBlock)))
	mux.Handle("GET /api/v1/instance/{name}/chat/contacts", requireInstance(http.HandlerFunc(h.HandleListContacts)))
	mux.Handle("GET /api/v1/instance/{name}/chat/profile/{jid}", requireInstance(http.HandlerFunc(h.HandleGetProfile)))

	// Group
	mux.Handle("POST /api/v1/instance/{name}/group/create", requireInstance(http.HandlerFunc(h.HandleCreateGroup)))
	mux.Handle("GET /api/v1/instance/{name}/group", requireInstance(http.HandlerFunc(h.HandleListGroups)))
	mux.Handle("GET /api/v1/instance/{name}/group/{jid}", requireInstance(http.HandlerFunc(h.HandleGetGroupInfo)))
	mux.Handle("PUT /api/v1/instance/{name}/group/{jid}/name", requireInstance(http.HandlerFunc(h.HandleUpdateGroupName)))
	mux.Handle("PUT /api/v1/instance/{name}/group/{jid}/description", requireInstance(http.HandlerFunc(h.HandleUpdateGroupDescription)))
	mux.Handle("PUT /api/v1/instance/{name}/group/{jid}/photo", requireInstance(http.HandlerFunc(h.HandleUpdateGroupPhoto)))
	mux.Handle("PUT /api/v1/instance/{name}/group/{jid}/settings", requireInstance(http.HandlerFunc(h.HandleUpdateGroupSettings)))
	mux.Handle("POST /api/v1/instance/{name}/group/{jid}/participants", requireInstance(http.HandlerFunc(h.HandleManageParticipants)))
	mux.Handle("GET /api/v1/instance/{name}/group/{jid}/invite-link", requireInstance(http.HandlerFunc(h.HandleGetInviteLink)))
	mux.Handle("DELETE /api/v1/instance/{name}/group/{jid}/leave", requireInstance(http.HandlerFunc(h.HandleLeaveGroup)))

	// Webhook
	mux.Handle("POST /api/v1/instance/{name}/webhook", requireInstance(http.HandlerFunc(h.HandleSetWebhook)))
	mux.Handle("GET /api/v1/instance/{name}/webhook", requireInstance(http.HandlerFunc(h.HandleGetWebhook)))
	mux.Handle("DELETE /api/v1/instance/{name}/webhook", requireInstance(http.HandlerFunc(h.HandleDeleteWebhook)))

	// Settings
	mux.Handle("GET /api/v1/instance/{name}/settings", requireInstance(http.HandlerFunc(h.HandleGetSettings)))
	mux.Handle("PUT /api/v1/instance/{name}/settings", requireInstance(http.HandlerFunc(h.HandleUpdateSettings)))
}

// --- Helpers ---

const maxRequestBody = 10 << 20 // 10MB

func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func readJSON(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBody)
	return json.NewDecoder(r.Body).Decode(v)
}
