package api

import (
	"net/http"
	"regexp"
	"time"

	"github.com/jholhewres/whatsgo-api/internal/auth"
	"github.com/jholhewres/whatsgo-api/internal/store"
)

var validInstanceName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

func (h *Handlers) HandleCreateInstance(w http.ResponseWriter, r *http.Request) {
	var req CreateInstanceRequest
	if err := readJSON(w, r, &req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.Name == "" {
		writeBadRequest(w, "name is required")
		return
	}
	if !validInstanceName.MatchString(req.Name) {
		writeBadRequest(w, "invalid instance name: must start with alphanumeric, contain only alphanumeric, hyphens and underscores, max 64 chars")
		return
	}

	inst, err := h.manager.Create(r.Context(), req.Name)
	if err != nil {
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "instance already exists", Code: "CONFLICT"})
		return
	}

	writeJSON(w, http.StatusCreated, CreateInstanceResponse{
		Instance: toInstanceInfo(inst, true),
	})
}

func (h *Handlers) HandleConnect(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req ConnectRequest
	readJSON(w, r, &req)

	waInst, err := h.manager.Connect(r.Context(), inst.Name)
	if err != nil {
		h.writeInternalError(w, r, "failed to connect instance", err)
		return
	}

	if req.PhoneNumber != "" {
		code, err := waInst.ConnectWithPairingCode(r.Context(), req.PhoneNumber)
		if err != nil {
			h.writeInternalError(w, r, "failed to generate pairing code", err)
			return
		}
		writeJSON(w, http.StatusOK, ConnectResponse{Status: "pairing_code", PairingCode: code})
		return
	}

	// QR code flow
	qrCh, unsub := waInst.SubscribeQR()
	defer unsub()

	select {
	case evt := <-qrCh:
		switch evt.Type {
		case "code":
			writeJSON(w, http.StatusOK, ConnectResponse{Status: "qr", QRCode: evt.Code})
		case "success":
			writeJSON(w, http.StatusOK, ConnectResponse{Status: "connected"})
		default:
			writeJSON(w, http.StatusOK, ConnectResponse{Status: evt.Type})
		}
	case <-time.After(30 * time.Second):
		writeJSON(w, http.StatusGatewayTimeout, ErrorResponse{Error: "QR code timeout", Code: "TIMEOUT"})
	case <-r.Context().Done():
		return
	}
}

func (h *Handlers) HandleRestart(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	if err := h.manager.Restart(r.Context(), inst.Name); err != nil {
		h.writeInternalError(w, r, "failed to restart instance", err)
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok", Message: "instance restarted"})
}

func (h *Handlers) HandleStatus(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	fresh, err := h.store.GetInstance(r.Context(), inst.ID)
	if err != nil {
		h.writeInternalError(w, r, "failed to get instance status", err)
		return
	}
	if fresh == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "instance not found", Code: "NOT_FOUND"})
		return
	}

	writeJSON(w, http.StatusOK, StatusResponse{Instance: toInstanceInfo(fresh, false)})
}

func (h *Handlers) HandleListInstances(w http.ResponseWriter, r *http.Request) {
	instances, err := h.store.ListInstances(r.Context())
	if err != nil {
		h.writeInternalError(w, r, "failed to list instances", err)
		return
	}

	list := make([]InstanceInfo, 0, len(instances))
	for _, inst := range instances {
		list = append(list, toInstanceInfo(inst, false))
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	if err := h.manager.Logout(r.Context(), inst.Name); err != nil {
		h.writeInternalError(w, r, "failed to logout instance", err)
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok", Message: "logged out"})
}

func (h *Handlers) HandleDeleteInstance(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	if err := h.manager.Delete(r.Context(), inst.Name); err != nil {
		h.writeInternalError(w, r, "failed to delete instance", err)
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok", Message: "instance deleted"})
}

func toInstanceInfo(inst *store.Instance, showToken bool) InstanceInfo {
	info := InstanceInfo{
		ID:           inst.ID,
		Name:         inst.Name,
		Status:       inst.Status,
		Phone:        inst.Phone,
		BusinessName: inst.BusinessName,
		CreatedAt:    inst.CreatedAt,
	}
	if showToken {
		info.Token = inst.Token
	}
	return info
}
