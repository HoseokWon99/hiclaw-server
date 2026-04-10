package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hiclaw-server/core"
)

// mockChatService implements core.ChatService
type mockChatService struct {
	sent    []*core.Message
	replies []*core.Message
}

func (m *mockChatService) SendMessage(sender *core.Device, content []*core.Part) (*core.Message, error) {
	msg := &core.Message{ID: "generated-id", Sender: sender, Content: content}
	m.sent = append(m.sent, msg)
	return msg, nil
}

func (m *mockChatService) HandleAgentReply(content []*core.Part) (*core.Message, error) {
	msg := &core.Message{ID: "agent-reply-id", Sender: &core.Device{Name: "agent"}, Content: content}
	m.replies = append(m.replies, msg)
	return msg, nil
}

func (m *mockChatService) ListRecent(limit int) ([]*core.Message, error) {
	return append(m.sent, m.replies...), nil
}

func TestPostMessage(t *testing.T) {
	svc := &mockChatService{}
	h := NewHandler(svc, nil)

	body := `{"content":[{"mime_type":"text/plain","data":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-IP", "100.64.0.1")
	req.Header.Set("X-Device-Name", "alice")
	w := httptest.NewRecorder()

	h.PostMessage(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	if len(svc.sent) != 1 {
		t.Error("expected 1 sent message")
	}
}

func TestAgentReply(t *testing.T) {
	svc := &mockChatService{}
	h := NewHandler(svc, nil)

	body := `{"content":[{"mime_type":"text/plain","data":"agent reply"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/agent/reply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAgentReply(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	if len(svc.replies) != 1 {
		t.Error("expected 1 agent reply")
	}
}

func TestGetMessages(t *testing.T) {
	svc := &mockChatService{}
	svc.SendMessage(&core.Device{Name: "a"}, []*core.Part{{Data: "hi"}})

	h := NewHandler(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/messages", nil)
	w := httptest.NewRecorder()

	h.GetMessages(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var msgs []*core.Message
	json.NewDecoder(w.Body).Decode(&msgs)
	if len(msgs) != 1 {
		t.Fatalf("expected 1, got %d", len(msgs))
	}
}
