package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hiclaw-server/core"
)

func TestClient_SendMessage(t *testing.T) {
	var received core.Message
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "/webhook/chat")
	msg := &core.Message{ID: "m1", Content: []*core.Part{{Data: "hello"}}, Timestamp: time.Now()}

	if err := c.SendMessage(msg); err != nil {
		t.Fatal(err)
	}
	if received.ID != "m1" {
		t.Errorf("expected m1, got %s", received.ID)
	}
}

func TestClient_Unreachable(t *testing.T) {
	c := NewClient("http://127.0.0.1:1", "/webhook/chat")
	err := c.SendMessage(&core.Message{ID: "m1"})
	if err == nil {
		t.Error("expected error for unreachable agent")
	}
}
