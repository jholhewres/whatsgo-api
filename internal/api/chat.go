package api

import (
	"net/http"

	"github.com/jholhewres/whatsgo-api/internal/auth"
)

func (h *Handlers) HandleCheckNumber(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req CheckNumberRequest
	if err := readJSON(r, &req); err != nil || len(req.Numbers) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "numbers array is required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	results, err := waInst.CheckNumbers(r.Context(), req.Numbers)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	checks := make([]NumberCheck, len(results))
	for idx, r := range results {
		checks[idx] = NumberCheck{Number: r.Number, Exists: r.Exists, JID: r.JID}
	}
	writeJSON(w, http.StatusOK, CheckNumberResponse{Results: checks})
}

func (h *Handlers) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req MarkReadRequest
	if err := readJSON(r, &req); err != nil || req.ChatJID == "" || len(req.MessageIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "chat_jid and message_ids are required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	if err := waInst.MarkRead(r.Context(), req.ChatJID, req.MessageIDs); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (h *Handlers) HandleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req DeleteMessageRequest
	if err := readJSON(r, &req); err != nil || req.ChatJID == "" || req.MessageID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "chat_jid and message_id are required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	if err := waInst.RevokeMessage(r.Context(), req.ChatJID, req.MessageID); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (h *Handlers) HandleEditMessage(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req EditMessageRequest
	if err := readJSON(r, &req); err != nil || req.ChatJID == "" || req.MessageID == "" || req.Text == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "chat_jid, message_id, and text are required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	resp, err := waInst.EditMessage(r.Context(), req.ChatJID, req.MessageID, req.Text)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SendResponse{
		MessageID: resp.MessageID,
		Status:    "edited",
		Timestamp: resp.Timestamp,
	})
}

func (h *Handlers) HandleSendPresence(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req SendPresenceRequest
	if err := readJSON(r, &req); err != nil || req.ChatJID == "" || req.Presence == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "chat_jid and presence are required"})
		return
	}

	switch req.Presence {
	case "composing", "paused", "recording":
	default:
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "presence must be one of: composing, paused, recording"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	if err := waInst.SendChatPresence(r.Context(), req.ChatJID, req.Presence); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (h *Handlers) HandleBlock(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req BlockRequest
	if err := readJSON(r, &req); err != nil || req.JID == "" || req.Action == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "jid and action are required"})
		return
	}

	if req.Action != "block" && req.Action != "unblock" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "action must be 'block' or 'unblock'"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	if err := waInst.UpdateBlocklist(r.Context(), req.JID, req.Action); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (h *Handlers) HandleListContacts(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	contacts, err := h.store.ListContacts(r.Context(), inst.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to list contacts"})
		return
	}

	writeJSON(w, http.StatusOK, contacts)
}

func (h *Handlers) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())
	jid := r.PathValue("jid")

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	profile, err := waInst.GetProfile(r.Context(), jid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, profile)
}
