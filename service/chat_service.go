package service

import (
	"log"
	"time"

	"hiclaw-server/core"
	"hiclaw-server/hub"

	"github.com/google/uuid"
)

type ChatServiceImpl struct {
	store        core.ChatStore
	hub          *hub.Hub
	agentService core.AgentService
}

func NewChatService(store core.ChatStore, h *hub.Hub, agentService core.AgentService) *ChatServiceImpl {
	return &ChatServiceImpl{store: store, hub: h, agentService: agentService}
}

func (s *ChatServiceImpl) SendMessage(sender *core.Device, content []*core.Part) (*core.Message, error) {
	msg := &core.Message{
		ID:        uuid.NewString(),
		Sender:    sender,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}

	if err := s.store.Save(msg); err != nil {
		return nil, err
	}

	if s.hub != nil {
		s.hub.Broadcast(msg)
	}

	if s.agentService != nil {
		go func() {
			if err := s.agentService.Notify(msg); err != nil {
				log.Printf("agent notify failed: %v", err)
			}
		}()
	}

	return msg, nil
}

func (s *ChatServiceImpl) HandleAgentReply(content []*core.Part) (*core.Message, error) {
	msg := &core.Message{
		ID:        uuid.NewString(),
		Sender:    &core.Device{Name: "agent"},
		Content:   content,
		Timestamp: time.Now().UTC(),
	}

	if err := s.store.Save(msg); err != nil {
		return nil, err
	}

	if s.hub != nil {
		s.hub.Broadcast(msg)
	}

	return msg, nil
}

func (s *ChatServiceImpl) ListRecent(limit int) ([]*core.Message, error) {
	return s.store.ListRecent(limit)
}
