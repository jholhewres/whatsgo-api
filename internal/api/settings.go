package api

import (
	"net/http"

	"github.com/jholhewres/whatsgo-api/internal/auth"
)

func (h *Handlers) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	settings, err := h.store.GetInstanceSettings(r.Context(), inst.ID)
	if err != nil {
		h.writeInternalError(w, r, "failed to get settings", err)
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

func (h *Handlers) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req UpdateSettingsRequest
	if err := readJSON(w, r, &req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}

	settings, err := h.store.GetInstanceSettings(r.Context(), inst.ID)
	if err != nil {
		h.writeInternalError(w, r, "failed to get settings", err)
		return
	}

	if req.RejectCall != nil {
		settings.RejectCall = *req.RejectCall
	}
	if req.RejectCallMessage != nil {
		settings.RejectCallMessage = *req.RejectCallMessage
	}
	if req.GroupsIgnore != nil {
		settings.GroupsIgnore = *req.GroupsIgnore
	}
	if req.AlwaysOnline != nil {
		settings.AlwaysOnline = *req.AlwaysOnline
	}
	if req.ReadMessages != nil {
		settings.ReadMessages = *req.ReadMessages
	}
	if req.ReadReceipts != nil {
		settings.ReadReceipts = *req.ReadReceipts
	}
	if req.WebhookBase64 != nil {
		settings.WebhookBase64 = *req.WebhookBase64
	}

	if err := h.store.UpsertInstanceSettings(r.Context(), settings); err != nil {
		h.writeInternalError(w, r, "failed to update settings", err)
		return
	}

	// Notify the running instance about settings change
	if waInst, ok := h.manager.Get(inst.Name); ok {
		waInst.UpdateSettings(settings)
	}

	writeJSON(w, http.StatusOK, settings)
}
