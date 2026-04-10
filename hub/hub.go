package hub

import (
	"log"
	"sync"

	"hiclaw-server/core"
)

type Conn interface {
	Send(msg *core.Message) error
}

type Hub struct {
	mu    sync.RWMutex
	conns map[string]Conn
}

func New() *Hub {
	return &Hub{conns: make(map[string]Conn)}
}

func (h *Hub) Add(deviceID string, conn Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[deviceID] = conn
}

func (h *Hub) Remove(deviceID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, deviceID)
}

func (h *Hub) Broadcast(msg *core.Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for id, conn := range h.conns {
		if err := conn.Send(msg); err != nil {
			log.Printf("send to %s failed: %v", id, err)
		}
	}
}

func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns)
}
