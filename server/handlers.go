package server

import (
	"encoding/json"
	"net/http"

	"hiclaw-server/core"
	"hiclaw-server/hub"
)

type Handler struct {
	chatSvc core.ChatService
	hub     *hub.Hub
}

func NewHandler(chatSvc core.ChatService, h *hub.Hub) *Handler {
	return &Handler{chatSvc: chatSvc, hub: h}
}

func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content []*core.Part `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	sender := &core.Device{
		IP:   r.Header.Get("X-Device-IP"),
		Name: r.Header.Get("X-Device-Name"),
	}

	msg, err := h.chatSvc.SendMessage(sender, req.Content)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func (h *Handler) HandleAgentReply(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content []*core.Part `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	msg, err := h.chatSvc.HandleAgentReply(req.Content)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	msgs, err := h.chatSvc.ListRecent(50)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}
