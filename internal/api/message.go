package api

import (
	"net/http"

	"github.com/jholhewres/whatsgo-api/internal/auth"
)

func (h *Handlers) HandleSendText(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req SendTextRequest
	if err := readJSON(w, r, &req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.Number == "" || req.Text == "" {
		writeBadRequest(w, "number and text are required")
		return
	}
	if !ValidatePhone(req.Number) {
		writeBadRequest(w, "invalid phone number format: use digits only, 7-15 characters")
		return
	}
	if len(req.Text) > MaxTextLength {
		writeBadRequest(w, "text exceeds maximum length of 65536 characters")
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeNotConnected(w)
		return
	}

	resp, err := waInst.SendText(r.Context(), req.Number, req.Text)
	if err != nil {
		h.writeInternalError(w, r, "failed to send text message", err)
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
	if err := readJSON(w, r, &req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.Number == "" || req.Media == "" || req.MediaType == "" {
		writeBadRequest(w, "number, media, and media_type are required")
		return
	}
	if !ValidatePhone(req.Number) {
		writeBadRequest(w, "invalid phone number format: use digits only, 7-15 characters")
		return
	}
	if !ValidateMediaType(req.MediaType) {
		writeBadRequest(w, "invalid media_type: must be one of image, video, audio, document")
		return
	}
	if req.Caption != "" && len(req.Caption) > MaxCaptionLength {
		writeBadRequest(w, "caption exceeds maximum length of 1024 characters")
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeNotConnected(w)
		return
	}

	resp, err := waInst.SendMedia(r.Context(), req.Number, req.MediaType, req.Media, req.Caption, req.Filename, req.MimeType)
	if err != nil {
		h.writeInternalError(w, r, "failed to send media", err)
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
	if err := readJSON(w, r, &req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.Number == "" {
		writeBadRequest(w, "number is required")
		return
	}
	if !ValidatePhone(req.Number) {
		writeBadRequest(w, "invalid phone number format: use digits only, 7-15 characters")
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 {
		writeBadRequest(w, "latitude must be between -90 and 90")
		return
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		writeBadRequest(w, "longitude must be between -180 and 180")
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeNotConnected(w)
		return
	}

	resp, err := waInst.SendLocation(r.Context(), req.Number, req.Latitude, req.Longitude, req.Name, req.Address)
	if err != nil {
		h.writeInternalError(w, r, "failed to send location", err)
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
	if err := readJSON(w, r, &req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.Number == "" || req.ContactName == "" || len(req.Phones) == 0 {
		writeBadRequest(w, "number, contact_name, and phones are required")
		return
	}
	if !ValidatePhone(req.Number) {
		writeBadRequest(w, "invalid phone number format: use digits only, 7-15 characters")
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeNotConnected(w)
		return
	}

	phones := make([]string, len(req.Phones))
	for i, p := range req.Phones {
		phones[i] = p.Number
	}

	resp, err := waInst.SendContact(r.Context(), req.Number, req.ContactName, phones)
	if err != nil {
		h.writeInternalError(w, r, "failed to send contact", err)
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
	if err := readJSON(w, r, &req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.Number == "" || req.MessageID == "" {
		writeBadRequest(w, "number and message_id are required")
		return
	}
	if !ValidatePhone(req.Number) {
		writeBadRequest(w, "invalid phone number format: use digits only, 7-15 characters")
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeNotConnected(w)
		return
	}

	resp, err := waInst.SendReaction(r.Context(), req.Number, req.MessageID, req.Emoji)
	if err != nil {
		h.writeInternalError(w, r, "failed to send reaction", err)
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
	if err := readJSON(w, r, &req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.Number == "" || req.Sticker == "" {
		writeBadRequest(w, "number and sticker are required")
		return
	}
	if !ValidatePhone(req.Number) {
		writeBadRequest(w, "invalid phone number format: use digits only, 7-15 characters")
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeNotConnected(w)
		return
	}

	resp, err := waInst.SendSticker(r.Context(), req.Number, req.Sticker, req.MimeType)
	if err != nil {
		h.writeInternalError(w, r, "failed to send sticker", err)
		return
	}

	writeJSON(w, http.StatusOK, SendResponse{
		MessageID: resp.MessageID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	})
}
