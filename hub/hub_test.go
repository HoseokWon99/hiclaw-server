package hub

import (
	"testing"
	"time"

	"hiclaw-server/core"
)

type fakeConn struct {
	received []*core.Message
}

func (f *fakeConn) Send(msg *core.Message) error {
	f.received = append(f.received, msg)
	return nil
}

func TestHub_Broadcast(t *testing.T) {
	h := New()
	c1 := &fakeConn{}
	c2 := &fakeConn{}
	h.Add("d1", c1)
	h.Add("d2", c2)

	msg := &core.Message{ID: "m1", Content: []*core.Part{{Data: "hi"}}, Timestamp: time.Now()}
	h.Broadcast(msg)

	if len(c1.received) != 1 || len(c2.received) != 1 {
		t.Errorf("expected both to receive 1 message")
	}
}

func TestHub_RemovePreventsReceive(t *testing.T) {
	h := New()
	c := &fakeConn{}
	h.Add("d1", c)
	h.Remove("d1")

	h.Broadcast(&core.Message{ID: "m1"})
	if len(c.received) != 0 {
		t.Error("removed conn should not receive")
	}
}

func TestHub_Count(t *testing.T) {
	h := New()
	h.Add("d1", &fakeConn{})
	h.Add("d2", &fakeConn{})
	if h.Count() != 2 {
		t.Errorf("expected 2, got %d", h.Count())
	}
}
