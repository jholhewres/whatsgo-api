package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/jholhewres/whatsgo-api/internal/auth"
	"github.com/jholhewres/whatsgo-api/internal/store"
)

func (h *Handlers) HandleSetWebhook(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req SetWebhookRequest
	if err := readJSON(w, r, &req); err != nil || req.URL == "" {
		writeBadRequest(w, "url is required")
		return
	}

	existing, err := h.store.GetWebhookByInstance(r.Context(), inst.ID)
	if err != nil {
		h.writeInternalError(w, r, "failed to check webhook", err)
		return
	}

	if existing != nil {
		existing.URL = req.URL
		existing.Events = req.Events
		existing.Headers = req.Headers
		existing.Enabled = req.Enabled
		if err := h.store.UpdateWebhook(r.Context(), existing); err != nil {
			h.writeInternalError(w, r, "failed to update webhook", err)
			return
		}
		writeJSON(w, http.StatusOK, existing)
		return
	}

	wh := &store.Webhook{
		ID:         uuid.New().String(),
		InstanceID: inst.ID,
		URL:        req.URL,
		Events:     req.Events,
		Headers:    req.Headers,
		Enabled:    req.Enabled,
		CreatedAt:  time.Now(),
	}
	if err := h.store.CreateWebhook(r.Context(), wh); err != nil {
		h.writeInternalError(w, r, "failed to create webhook", err)
		return
	}

	writeJSON(w, http.StatusCreated, wh)
}

func (h *Handlers) HandleGetWebhook(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	wh, err := h.store.GetWebhookByInstance(r.Context(), inst.ID)
	if err != nil {
		h.writeInternalError(w, r, "failed to get webhook", err)
		return
	}
	if wh == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "no webhook configured", Code: "NOT_FOUND"})
		return
	}

	writeJSON(w, http.StatusOK, wh)
}

func (h *Handlers) HandleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	if err := h.store.DeleteWebhook(r.Context(), inst.ID); err != nil {
		h.writeInternalError(w, r, "failed to delete webhook", err)
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok", Message: "webhook deleted"})
}
