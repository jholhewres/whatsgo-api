package api

import (
	"net/http"

	"github.com/jholhewres/whatsgo-api/internal/auth"
)

func (h *Handlers) HandleSendText(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req SendTextRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.Number == "" || req.Text == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "number and text are required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	resp, err := waInst.SendText(r.Context(), req.Number, req.Text)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SendResponse{
		MessageID: resp.MessageID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	})
}

func (h *Handlers) HandleSendMedia(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req SendMediaRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.Number == "" || req.Media == "" || req.MediaType == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "number, media, and media_type are required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	resp, err := waInst.SendMedia(r.Context(), req.Number, req.MediaType, req.Media, req.Caption, req.Filename, req.MimeType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SendResponse{
		MessageID: resp.MessageID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	})
}

func (h *Handlers) HandleSendLocation(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req SendLocationRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.Number == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "number is required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	resp, err := waInst.SendLocation(r.Context(), req.Number, req.Latitude, req.Longitude, req.Name, req.Address)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SendResponse{
		MessageID: resp.MessageID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	})
}

func (h *Handlers) HandleSendContact(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req SendContactRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.Number == "" || req.ContactName == "" || len(req.Phones) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "number, contact_name, and phones are required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	phones := make([]string, len(req.Phones))
	for i, p := range req.Phones {
		phones[i] = p.Number
	}

	resp, err := waInst.SendContact(r.Context(), req.Number, req.ContactName, phones)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SendResponse{
		MessageID: resp.MessageID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	})
}

func (h *Handlers) HandleSendReaction(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req SendReactionRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.Number == "" || req.MessageID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "number and message_id are required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	resp, err := waInst.SendReaction(r.Context(), req.Number, req.MessageID, req.Emoji)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SendResponse{
		MessageID: resp.MessageID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	})
}

func (h *Handlers) HandleSendSticker(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req SendStickerRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.Number == "" || req.Sticker == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "number and sticker are required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	resp, err := waInst.SendSticker(r.Context(), req.Number, req.Sticker, req.MimeType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SendResponse{
		MessageID: resp.MessageID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	})
}
