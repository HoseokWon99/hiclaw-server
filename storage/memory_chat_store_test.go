package storage

import (
	"fmt"
	"testing"
	"time"

	"hiclaw-server/core"
)

var _ core.ChatStore = (*MemoryChatStore)(nil)

func TestMemoryChatStore_SaveAndList(t *testing.T) {
	store := NewMemoryChatStore()

	msg := &core.Message{
		ID:        "msg-1",
		Content:   []*core.Part{{MimeType: "text/plain", Data: "hello"}},
		Timestamp: time.Now(),
	}
	if err := store.Save(msg); err != nil {
		t.Fatal(err)
	}

	msgs, err := store.ListRecent(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].ID != "msg-1" {
		t.Fatalf("expected [msg-1], got %v", msgs)
	}
}

func TestMemoryChatStore_ListRecentLimit(t *testing.T) {
	store := NewMemoryChatStore()
	for i := 0; i < 20; i++ {
		store.Save(&core.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			Content:   []*core.Part{{MimeType: "text/plain", Data: "hi"}},
			Timestamp: time.Now(),
		})
	}

	msgs, _ := store.ListRecent(5)
	if len(msgs) != 5 {
		t.Fatalf("expected 5, got %d", len(msgs))
	}
	if msgs[0].ID != "msg-19" {
		t.Errorf("expected most recent first (msg-19), got %s", msgs[0].ID)
	}
}
