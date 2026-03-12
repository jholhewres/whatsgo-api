package api

import (
	"io"
	"net/http"

	"github.com/jholhewres/whatsgo-api/internal/auth"
)

func (h *Handlers) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	var req CreateGroupRequest
	if err := readJSON(r, &req); err != nil || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "name is required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	resp, err := waInst.CreateGroup(r.Context(), req.Name, req.Participants)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handlers) HandleListGroups(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	groups, err := waInst.GetJoinedGroups(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, groups)
}

func (h *Handlers) HandleGetGroupInfo(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())
	jid := r.PathValue("jid")

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	info, err := waInst.GetGroupInfo(r.Context(), jid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, info)
}

func (h *Handlers) HandleUpdateGroupName(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())
	jid := r.PathValue("jid")

	var req UpdateGroupNameRequest
	if err := readJSON(r, &req); err != nil || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "name is required"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	if err := waInst.SetGroupName(r.Context(), jid, req.Name); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (h *Handlers) HandleUpdateGroupDescription(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())
	jid := r.PathValue("jid")

	var req UpdateGroupDescriptionRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	if err := waInst.SetGroupDescription(r.Context(), jid, req.Description); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (h *Handlers) HandleUpdateGroupPhoto(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())
	jid := r.PathValue("jid")

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	data, err := readMediaFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "failed to read photo data"})
		return
	}

	if err := waInst.SetGroupPhoto(r.Context(), jid, data); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (h *Handlers) HandleUpdateGroupSettings(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())
	jid := r.PathValue("jid")

	var req UpdateGroupSettingsRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	if err := waInst.UpdateGroupSettings(r.Context(), jid, req.Locked, req.Announce); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (h *Handlers) HandleManageParticipants(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())
	jid := r.PathValue("jid")

	var req ManageParticipantsRequest
	if err := readJSON(r, &req); err != nil || req.Action == "" || len(req.Participants) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "action and participants are required"})
		return
	}

	switch req.Action {
	case "add", "remove", "promote", "demote":
	default:
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "action must be one of: add, remove, promote, demote"})
		return
	}

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	if err := waInst.ManageParticipants(r.Context(), jid, req.Action, req.Participants); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (h *Handlers) HandleGetInviteLink(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())
	jid := r.PathValue("jid")

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	link, err := waInst.GetGroupInviteLink(r.Context(), jid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"invite_link": link})
}

func (h *Handlers) HandleLeaveGroup(w http.ResponseWriter, r *http.Request) {
	inst := auth.GetInstance(r.Context())
	jid := r.PathValue("jid")

	waInst, ok := h.manager.Get(inst.Name)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "instance not connected"})
		return
	}

	if err := waInst.LeaveGroup(r.Context(), jid); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func readMediaFromRequest(r *http.Request) ([]byte, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, 10<<20) // 10MB max

	file, _, err := r.FormFile("photo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
