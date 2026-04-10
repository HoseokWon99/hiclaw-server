package service

import (
	"testing"

	"hiclaw-server/core"
	"hiclaw-server/hub"
)

var _ core.ChatService = (*ChatServiceImpl)(nil)

// mockChatStore implements core.ChatStore
type mockChatStore struct {
	messages []*core.Message
}

func (m *mockChatStore) Save(msg *core.Message) error {
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockChatStore) ListRecent(limit int) ([]*core.Message, error) {
	if limit > len(m.messages) {
		limit = len(m.messages)
	}
	return m.messages[len(m.messages)-limit:], nil
}

type fakeHubConn struct {
	received []*core.Message
}

func (f *fakeHubConn) Send(msg *core.Message) error {
	f.received = append(f.received, msg)
	return nil
}

func TestChatService_SendMessage(t *testing.T) {
	store := &mockChatStore{}
	h := hub.New()
	conn := &fakeHubConn{}
	h.Add("user-1", conn)

	svc := NewChatService(store, h, nil)

	msg, err := svc.SendMessage(
		&core.Device{IP: "100.64.0.1", Name: "alice"},
		[]*core.Part{{MimeType: "text/plain", Data: "hello"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if msg.ID == "" {
		t.Error("expected ID to be assigned")
	}
	if len(store.messages) != 1 {
		t.Error("expected message stored")
	}
	if len(conn.received) != 1 {
		t.Error("expected message broadcast")
	}
}

func TestChatService_HandleAgentReply(t *testing.T) {
	store := &mockChatStore{}
	h := hub.New()
	conn := &fakeHubConn{}
	h.Add("user-1", conn)

	svc := NewChatService(store, h, nil)

	msg, err := svc.HandleAgentReply([]*core.Part{{Data: "agent says hi"}})
	if err != nil {
		t.Fatal(err)
	}
	if msg.Sender.Name != "agent" {
		t.Errorf("expected sender 'agent', got %s", msg.Sender.Name)
	}
	if len(conn.received) != 1 {
		t.Error("expected broadcast")
	}
}

func TestChatService_ListRecent(t *testing.T) {
	store := &mockChatStore{}
	svc := NewChatService(store, nil, nil)

	store.Save(&core.Message{ID: "m1"})
	msgs, _ := svc.ListRecent(10)
	if len(msgs) != 1 {
		t.Fatalf("expected 1, got %d", len(msgs))
	}
}
