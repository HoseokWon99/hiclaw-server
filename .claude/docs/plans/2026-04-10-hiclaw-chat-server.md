# Hiclaw Chat Server Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a chat server that communicates with openclaw agent devices over Tailscale VPN, broadcasting messages between user devices and an AI agent.

**Architecture:** Clean architecture with interfaces defined in `core/` package. Storage and service implementations depend on `core` interfaces — never the reverse. The `core` package contains domain types, storage interfaces (`ChatStore`, `DeviceStore`), and service interfaces (`ChatService`, `DeviceService`). Implementations live in `storage/` and `service/`. HTTP handlers depend only on service interfaces.

**Tech Stack:** Go 1.26, Tailscale SDK (`tailscale.com`), `coder/websocket`, CouchBase Go SDK, `modernc.org/sqlite`, `net/http`

**Package dependency graph:**
```
main → server → core (interfaces)
              → service (implements core.ChatService, core.DeviceService)
                  → core
       storage (implements core.ChatStore, core.DeviceStore)
                  → core
       hub → core
       agent → core
       tailnet → core
       config
```

---

### Task 1: Core Types & Interfaces

**Files:**
- Modify: `core/types.go`
- Create: `core/stores.go`
- Create: `core/services.go`

**Step 1: Write updated core types**

Replace `core/types.go` — use string IP for JSON friendliness:

```go
// core/types.go
package core

import "time"

type Device struct {
	IP      string `json:"ip"`
	Name    string `json:"name"`
	IsAgent bool   `json:"is_agent"`
}

type Part struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type Message struct {
	ID        string    `json:"id"`
	Sender    *Device   `json:"sender"`
	Content   []*Part   `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
```

**Step 2: Write storage interfaces**

```go
// core/stores.go
package core

type ChatStore interface {
	Save(msg *Message) error
	ListRecent(limit int) ([]*Message, error)
}

type DeviceStore interface {
	Register(device *Device) error
	Remove(ip string) error
	ListAll() ([]*Device, error)
	GetAgent() (*Device, error)
}
```

**Step 3: Write service interfaces**

```go
// core/services.go
package core

type ChatService interface {
	SendMessage(sender *Device, content []*Part) (*Message, error)
	HandleAgentReply(content []*Part) (*Message, error)
	ListRecent(limit int) ([]*Message, error)
}

type DeviceService interface {
	Register(device *Device) error
	Remove(ip string) error
	ListAll() ([]*Device, error)
	GetAgent() (*Device, error)
}
```

**Step 4: Verify it compiles**

Run: `go build ./core/`
Expected: compiles

**Step 5: Commit**

```bash
git add core/
git commit -m "feat: define domain types and interfaces for stores and services in core"
```

---

### Task 2: Configuration

**Files:**
- Create: `config/config.go`
- Create: `config/config_test.go`

**Step 1: Write the failing test**

```go
// config/config_test.go
package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := Load()
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("expected :8080, got %s", cfg.HTTPAddr)
	}
	if cfg.SQLitePath != "devices.db" {
		t.Errorf("expected devices.db, got %s", cfg.SQLitePath)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("HICLAW_HTTP_ADDR", ":9090")
	defer os.Unsetenv("HICLAW_HTTP_ADDR")

	cfg := Load()
	if cfg.HTTPAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.HTTPAddr)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./config/ -v`
Expected: FAIL — package does not exist

**Step 3: Write minimal implementation**

```go
// config/config.go
package config

import "os"

type Config struct {
	HTTPAddr         string
	CouchbaseURL     string
	CouchbaseBucket  string
	CouchbaseUser    string
	CouchbasePass    string
	SQLitePath       string
	AgentWebhookPath string
	AgentDeviceName  string
}

func Load() Config {
	cfg := Config{
		HTTPAddr:         ":8080",
		CouchbaseURL:     "couchbase://localhost",
		CouchbaseBucket:  "hiclaw",
		SQLitePath:       "devices.db",
		AgentWebhookPath: "/webhook/chat",
		AgentDeviceName:  "openclaw-agent",
	}
	if v := os.Getenv("HICLAW_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := os.Getenv("HICLAW_COUCHBASE_URL"); v != "" {
		cfg.CouchbaseURL = v
	}
	if v := os.Getenv("HICLAW_COUCHBASE_BUCKET"); v != "" {
		cfg.CouchbaseBucket = v
	}
	if v := os.Getenv("HICLAW_COUCHBASE_USER"); v != "" {
		cfg.CouchbaseUser = v
	}
	if v := os.Getenv("HICLAW_COUCHBASE_PASS"); v != "" {
		cfg.CouchbasePass = v
	}
	if v := os.Getenv("HICLAW_SQLITE_PATH"); v != "" {
		cfg.SQLitePath = v
	}
	if v := os.Getenv("HICLAW_AGENT_WEBHOOK_PATH"); v != "" {
		cfg.AgentWebhookPath = v
	}
	if v := os.Getenv("HICLAW_AGENT_DEVICE_NAME"); v != "" {
		cfg.AgentDeviceName = v
	}
	return cfg
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./config/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add config/
git commit -m "feat: add env-based configuration"
```

---

### Task 3: SQLite Device Store (implements `core.DeviceStore`)

**Files:**
- Create: `storage/sqlite_device_store.go`
- Create: `storage/sqlite_device_store_test.go`

**Step 1: Write the failing test**

```go
// storage/sqlite_device_store_test.go
package storage

import (
	"testing"

	"hiclaw-server/core"
)

// Compile-time interface check
var _ core.DeviceStore = (*SQLiteDeviceStore)(nil)

func TestSQLiteDeviceStore_RegisterAndList(t *testing.T) {
	store, err := NewSQLiteDeviceStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	err = store.Register(&core.Device{IP: "100.64.0.1", Name: "alice-laptop", IsAgent: false})
	if err != nil {
		t.Fatal(err)
	}
	err = store.Register(&core.Device{IP: "100.64.0.2", Name: "openclaw-agent", IsAgent: true})
	if err != nil {
		t.Fatal(err)
	}

	devices, err := store.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
}

func TestSQLiteDeviceStore_Upsert(t *testing.T) {
	store, err := NewSQLiteDeviceStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	store.Register(&core.Device{IP: "100.64.0.1", Name: "alice-laptop"})
	store.Register(&core.Device{IP: "100.64.0.1", Name: "alice-desktop"})

	devices, _ := store.ListAll()
	if len(devices) != 1 {
		t.Fatalf("expected 1, got %d", len(devices))
	}
	if devices[0].Name != "alice-desktop" {
		t.Errorf("expected alice-desktop, got %s", devices[0].Name)
	}
}

func TestSQLiteDeviceStore_Remove(t *testing.T) {
	store, err := NewSQLiteDeviceStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	store.Register(&core.Device{IP: "100.64.0.1", Name: "alice"})
	store.Remove("100.64.0.1")

	devices, _ := store.ListAll()
	if len(devices) != 0 {
		t.Fatalf("expected 0, got %d", len(devices))
	}
}

func TestSQLiteDeviceStore_GetAgent(t *testing.T) {
	store, err := NewSQLiteDeviceStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	store.Register(&core.Device{IP: "100.64.0.1", Name: "alice", IsAgent: false})
	store.Register(&core.Device{IP: "100.64.0.2", Name: "openclaw-agent", IsAgent: true})

	agent, err := store.GetAgent()
	if err != nil {
		t.Fatal(err)
	}
	if agent.IP != "100.64.0.2" {
		t.Errorf("expected 100.64.0.2, got %s", agent.IP)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./storage/ -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// storage/sqlite_device_store.go
package storage

import (
	"database/sql"
	"fmt"

	"hiclaw-server/core"

	_ "modernc.org/sqlite"
)

type SQLiteDeviceStore struct {
	db *sql.DB
}

func NewSQLiteDeviceStore(path string) (*SQLiteDeviceStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS devices (
		ip TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		is_agent BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}
	return &SQLiteDeviceStore{db: db}, nil
}

func (s *SQLiteDeviceStore) Register(d *core.Device) error {
	_, err := s.db.Exec(
		`INSERT INTO devices (ip, name, is_agent) VALUES (?, ?, ?)
		 ON CONFLICT(ip) DO UPDATE SET name = excluded.name, is_agent = excluded.is_agent`,
		d.IP, d.Name, d.IsAgent,
	)
	return err
}

func (s *SQLiteDeviceStore) Remove(ip string) error {
	_, err := s.db.Exec(`DELETE FROM devices WHERE ip = ?`, ip)
	return err
}

func (s *SQLiteDeviceStore) ListAll() ([]*core.Device, error) {
	rows, err := s.db.Query(`SELECT ip, name, is_agent FROM devices`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*core.Device
	for rows.Next() {
		d := &core.Device{}
		if err := rows.Scan(&d.IP, &d.Name, &d.IsAgent); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (s *SQLiteDeviceStore) GetAgent() (*core.Device, error) {
	d := &core.Device{}
	err := s.db.QueryRow(`SELECT ip, name, is_agent FROM devices WHERE is_agent = TRUE LIMIT 1`).
		Scan(&d.IP, &d.Name, &d.IsAgent)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *SQLiteDeviceStore) Close() error {
	return s.db.Close()
}
```

**Step 4: Run test to verify it passes**

Run: `go get modernc.org/sqlite && go test ./storage/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add storage/ go.mod go.sum
git commit -m "feat: add SQLite device store implementing core.DeviceStore"
```

---

### Task 4: In-Memory Chat Store (implements `core.ChatStore`)

**Files:**
- Create: `storage/memory_chat_store.go`
- Create: `storage/memory_chat_store_test.go`

**Step 1: Write the failing test**

```go
// storage/memory_chat_store_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./storage/ -v -run MemoryChat`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// storage/memory_chat_store.go
package storage

import (
	"sync"

	"hiclaw-server/core"
)

type MemoryChatStore struct {
	mu       sync.Mutex
	messages []*core.Message
}

func NewMemoryChatStore() *MemoryChatStore {
	return &MemoryChatStore{}
}

func (s *MemoryChatStore) Save(msg *core.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
	return nil
}

func (s *MemoryChatStore) ListRecent(limit int) ([]*core.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := len(s.messages)
	if limit > n {
		limit = n
	}
	result := make([]*core.Message, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.messages[n-1-i]
	}
	return result, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./storage/ -v -run MemoryChat`
Expected: PASS

**Step 5: Commit**

```bash
git add storage/memory_chat_store.go storage/memory_chat_store_test.go
git commit -m "feat: add in-memory chat store implementing core.ChatStore"
```

---

### Task 5: WebSocket Hub

**Files:**
- Create: `hub/hub.go`
- Create: `hub/hub_test.go`

**Step 1: Write the failing test**

```go
// hub/hub_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./hub/ -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// hub/hub.go
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./hub/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add hub/
git commit -m "feat: add WebSocket hub for connection management and broadcasting"
```

---

### Task 6: Agent Webhook Client

**Files:**
- Create: `agent/client.go`
- Create: `agent/client_test.go`

**Step 1: Write the failing test**

```go
// agent/client_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./agent/ -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// agent/client.go
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hiclaw-server/core"
)

type Client struct {
	baseURL     string
	webhookPath string
	http        *http.Client
}

func NewClient(baseURL, webhookPath string) *Client {
	return &Client{
		baseURL:     baseURL,
		webhookPath: webhookPath,
		http:        &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) SendMessage(msg *core.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	resp, err := c.http.Post(c.baseURL+c.webhookPath, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send to agent: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("agent returned %d", resp.StatusCode)
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./agent/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/
git commit -m "feat: add agent webhook client"
```

---

### Task 7: Device Service (implements `core.DeviceService`)

**Files:**
- Create: `service/device_service.go`
- Create: `service/device_service_test.go`

**Step 1: Write the failing test**

```go
// service/device_service_test.go
package service

import (
	"fmt"
	"testing"

	"hiclaw-server/core"
)

var _ core.DeviceService = (*DeviceServiceImpl)(nil)

// mockDeviceStore implements core.DeviceStore for testing
type mockDeviceStore struct {
	devices []*core.Device
}

func (m *mockDeviceStore) Register(d *core.Device) error {
	for i, existing := range m.devices {
		if existing.IP == d.IP {
			m.devices[i] = d
			return nil
		}
	}
	m.devices = append(m.devices, d)
	return nil
}

func (m *mockDeviceStore) Remove(ip string) error {
	for i, d := range m.devices {
		if d.IP == ip {
			m.devices = append(m.devices[:i], m.devices[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockDeviceStore) ListAll() ([]*core.Device, error) {
	return m.devices, nil
}

func (m *mockDeviceStore) GetAgent() (*core.Device, error) {
	for _, d := range m.devices {
		if d.IsAgent {
			return d, nil
		}
	}
	return nil, fmt.Errorf("no agent found")
}

func TestDeviceService_RegisterAndList(t *testing.T) {
	store := &mockDeviceStore{}
	svc := NewDeviceService(store)

	err := svc.Register(&core.Device{IP: "100.64.0.1", Name: "alice"})
	if err != nil {
		t.Fatal(err)
	}

	devices, err := svc.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1, got %d", len(devices))
	}
}

func TestDeviceService_GetAgent(t *testing.T) {
	store := &mockDeviceStore{}
	svc := NewDeviceService(store)

	svc.Register(&core.Device{IP: "100.64.0.1", Name: "alice"})
	svc.Register(&core.Device{IP: "100.64.0.2", Name: "agent", IsAgent: true})

	agent, err := svc.GetAgent()
	if err != nil {
		t.Fatal(err)
	}
	if agent.Name != "agent" {
		t.Errorf("expected agent, got %s", agent.Name)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./service/ -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// service/device_service.go
package service

import "hiclaw-server/core"

type DeviceServiceImpl struct {
	store core.DeviceStore
}

func NewDeviceService(store core.DeviceStore) *DeviceServiceImpl {
	return &DeviceServiceImpl{store: store}
}

func (s *DeviceServiceImpl) Register(d *core.Device) error {
	return s.store.Register(d)
}

func (s *DeviceServiceImpl) Remove(ip string) error {
	return s.store.Remove(ip)
}

func (s *DeviceServiceImpl) ListAll() ([]*core.Device, error) {
	return s.store.ListAll()
}

func (s *DeviceServiceImpl) GetAgent() (*core.Device, error) {
	return s.store.GetAgent()
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./service/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add service/
git commit -m "feat: add DeviceService implementing core.DeviceService"
```

---

### Task 8: Chat Service (implements `core.ChatService`)

**Files:**
- Create: `service/chat_service.go`
- Create: `service/chat_service_test.go`

**Step 1: Write the failing test**

```go
// service/chat_service_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./service/ -v -run ChatService`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// service/chat_service.go
package service

import (
	"log"
	"time"

	"hiclaw-server/agent"
	"hiclaw-server/core"
	"hiclaw-server/hub"

	"github.com/google/uuid"
)

type ChatServiceImpl struct {
	store       core.ChatStore
	hub         *hub.Hub
	agentClient *agent.Client
}

func NewChatService(store core.ChatStore, h *hub.Hub, ac *agent.Client) *ChatServiceImpl {
	return &ChatServiceImpl{store: store, hub: h, agentClient: ac}
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

	if s.agentClient != nil {
		go func() {
			if err := s.agentClient.SendMessage(msg); err != nil {
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./service/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add service/chat_service.go service/chat_service_test.go
git commit -m "feat: add ChatService implementing core.ChatService with broadcast and agent forwarding"
```

---

### Task 9: HTTP Handlers (depend on `core.ChatService` interface)

**Files:**
- Create: `server/handlers.go`
- Create: `server/handlers_test.go`

**Step 1: Write the failing test**

```go
// server/handlers_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./server/ -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// server/handlers.go
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./server/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/
git commit -m "feat: add HTTP handlers depending on core.ChatService interface"
```

---

### Task 10: WebSocket Handler

**Files:**
- Create: `server/ws_conn.go`
- Modify: `server/handlers.go` — add `HandleWebSocket`

**Step 1: Write WebSocket connection wrapper**

```go
// server/ws_conn.go
package server

import (
	"context"
	"encoding/json"

	"hiclaw-server/core"

	"github.com/coder/websocket"
)

type WSConn struct {
	conn *websocket.Conn
}

func NewWSConn(conn *websocket.Conn) *WSConn {
	return &WSConn{conn: conn}
}

func (w *WSConn) Send(msg *core.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return w.conn.Write(context.Background(), websocket.MessageText, data)
}

func (w *WSConn) ReadMessage() (*core.Message, error) {
	_, data, err := w.conn.Read(context.Background())
	if err != nil {
		return nil, err
	}
	var msg core.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (w *WSConn) Close() error {
	return w.conn.CloseNow()
}
```

**Step 2: Add HandleWebSocket to handlers.go**

Append to `server/handlers.go`:

```go
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("ws accept: %v", err)
		return
	}

	deviceIP := r.Header.Get("X-Device-IP")
	deviceName := r.Header.Get("X-Device-Name")
	if deviceIP == "" {
		deviceIP = r.RemoteAddr
	}

	wsConn := NewWSConn(conn)
	h.hub.Add(deviceIP, wsConn)
	defer func() {
		h.hub.Remove(deviceIP)
		wsConn.Close()
	}()

	log.Printf("ws connected: %s (%s)", deviceName, deviceIP)

	for {
		msg, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("ws disconnected: %s (%v)", deviceIP, err)
			return
		}
		h.chatSvc.SendMessage(
			&core.Device{IP: deviceIP, Name: deviceName},
			msg.Content,
		)
	}
}
```

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add server/
git commit -m "feat: add WebSocket handler with hub integration"
```

---

### Task 11: Tailscale Device Discovery

**Files:**
- Create: `tailnet/discovery.go`

**Step 1: Implement discovery using `core.DeviceService` interface**

```go
// tailnet/discovery.go
package tailnet

import (
	"context"
	"log"
	"time"

	"hiclaw-server/core"

	"tailscale.com/client/tailscale"
)

type Discovery struct {
	client    tailscale.LocalClient
	deviceSvc core.DeviceService
	agentName string
}

func NewDiscovery(deviceSvc core.DeviceService, agentName string) *Discovery {
	return &Discovery{deviceSvc: deviceSvc, agentName: agentName}
}

func (d *Discovery) SyncOnce(ctx context.Context) error {
	status, err := d.client.Status(ctx)
	if err != nil {
		return err
	}
	for _, peer := range status.Peer {
		if len(peer.TailscaleIPs) == 0 {
			continue
		}
		device := &core.Device{
			IP:      peer.TailscaleIPs[0].String(),
			Name:    peer.HostName,
			IsAgent: peer.HostName == d.agentName,
		}
		if err := d.deviceSvc.Register(device); err != nil {
			log.Printf("register %s failed: %v", device.Name, err)
		}
	}
	return nil
}

func (d *Discovery) RunLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.SyncOnce(ctx); err != nil {
				log.Printf("tailnet sync: %v", err)
			}
		}
	}
}
```

**Step 2: Commit**

```bash
git add tailnet/
git commit -m "feat: add Tailscale discovery using core.DeviceService interface"
```

---

### Task 12: Wire Everything in main.go

**Files:**
- Modify: `main.go`

**Step 1: Replace boilerplate with full wiring**

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hiclaw-server/agent"
	"hiclaw-server/config"
	"hiclaw-server/hub"
	"hiclaw-server/server"
	"hiclaw-server/service"
	"hiclaw-server/storage"
	"hiclaw-server/tailnet"
)

func main() {
	cfg := config.Load()

	// Storage layer (implements core interfaces)
	deviceStore, err := storage.NewSQLiteDeviceStore(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("device store: %v", err)
	}
	defer deviceStore.Close()

	chatStore := storage.NewMemoryChatStore()

	// Infrastructure
	h := hub.New()
	agentClient := agent.NewClient("http://"+cfg.AgentDeviceName, cfg.AgentWebhookPath)

	// Service layer (implements core interfaces, depends on core store interfaces)
	deviceSvc := service.NewDeviceService(deviceStore)
	chatSvc := service.NewChatService(chatStore, h, agentClient)

	// Tailscale discovery (depends on core.DeviceService)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	disc := tailnet.NewDiscovery(deviceSvc, cfg.AgentDeviceName)
	if err := disc.SyncOnce(ctx); err != nil {
		log.Printf("initial tailnet sync (continuing): %v", err)
	}
	go disc.RunLoop(ctx, 30*time.Second)

	// HTTP layer (depends on core.ChatService)
	handler := server.NewHandler(chatSvc, h)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/messages", handler.GetMessages)
	mux.HandleFunc("POST /api/messages", handler.PostMessage)
	mux.HandleFunc("POST /api/agent/reply", handler.HandleAgentReply)
	mux.HandleFunc("GET /ws", handler.HandleWebSocket)

	log.Printf("hiclaw-server starting on %s", cfg.HTTPAddr)
	go func() {
		if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
}
```

**Step 2: Build**

Run: `go build ./...`
Expected: compiles

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: wire all layers in main with dependency injection via interfaces"
```

---

### Task 13: CouchBase Chat Store (implements `core.ChatStore`)

**Files:**
- Create: `storage/couchbase_chat_store.go`

Swap in `main.go` when CouchBase credentials are configured. Same `core.ChatStore` interface — zero changes to service or handler layers.

**Step 1: Implement**

```go
// storage/couchbase_chat_store.go
package storage

import (
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
	"hiclaw-server/core"
)

type CouchBaseChatStore struct {
	cluster *gocb.Cluster
	col     *gocb.Collection
	bucket  string
}

func NewCouchBaseChatStore(url, bucket, user, pass string) (*CouchBaseChatStore, error) {
	cluster, err := gocb.Connect(url, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{Username: user, Password: pass},
	})
	if err != nil {
		return nil, fmt.Errorf("couchbase connect: %w", err)
	}
	if err := cluster.WaitUntilReady(5*time.Second, nil); err != nil {
		return nil, fmt.Errorf("couchbase not ready: %w", err)
	}
	b := cluster.Bucket(bucket)
	return &CouchBaseChatStore{cluster: cluster, col: b.DefaultCollection(), bucket: bucket}, nil
}

func (s *CouchBaseChatStore) Save(msg *core.Message) error {
	_, err := s.col.Upsert(msg.ID, msg, nil)
	return err
}

func (s *CouchBaseChatStore) ListRecent(limit int) ([]*core.Message, error) {
	query := fmt.Sprintf(
		"SELECT META().id, m.* FROM `%s` m ORDER BY m.timestamp DESC LIMIT $1",
		s.bucket,
	)
	result, err := s.cluster.Query(query, &gocb.QueryOptions{
		PositionalParameters: []interface{}{limit},
	})
	if err != nil {
		return nil, err
	}
	var messages []*core.Message
	for result.Next() {
		var msg core.Message
		if err := result.Row(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, result.Err()
}
```

**Step 2: Update main.go to conditionally use CouchBase**

```go
// In main.go, replace chatStore init:
var chatStore core.ChatStore
if cfg.CouchbaseUser != "" {
	cs, err := storage.NewCouchBaseChatStore(cfg.CouchbaseURL, cfg.CouchbaseBucket, cfg.CouchbaseUser, cfg.CouchbasePass)
	if err != nil {
		log.Fatalf("couchbase: %v", err)
	}
	chatStore = cs
	log.Println("using CouchBase chat store")
} else {
	chatStore = storage.NewMemoryChatStore()
	log.Println("using in-memory chat store")
}
```

**Step 3: Commit**

```bash
git add storage/couchbase_chat_store.go main.go go.mod go.sum
git commit -m "feat: add CouchBase chat store implementing core.ChatStore"
```

---

## Summary

| # | Package | Component | Implements |
|---|---------|-----------|------------|
| 1 | `core` | Types + Interfaces | `ChatStore`, `DeviceStore`, `ChatService`, `DeviceService` |
| 2 | `config` | Configuration | — |
| 3 | `storage` | SQLite device store | `core.DeviceStore` |
| 4 | `storage` | In-memory chat store | `core.ChatStore` |
| 5 | `hub` | WebSocket hub | — |
| 6 | `agent` | Webhook client | — |
| 7 | `service` | Device service | `core.DeviceService` (depends on `core.DeviceStore`) |
| 8 | `service` | Chat service | `core.ChatService` (depends on `core.ChatStore`, hub, agent) |
| 9 | `server` | HTTP handlers | depends on `core.ChatService` |
| 10 | `server` | WebSocket handler | depends on `core.ChatService` + hub |
| 11 | `tailnet` | Tailscale discovery | depends on `core.DeviceService` |
| 12 | `main` | Wire all layers | DI via interfaces |
| 13 | `storage` | CouchBase chat store | `core.ChatStore` |

**Dependency direction:** `main` → `server`/`service`/`storage`/`tailnet` → `core` (interfaces only). No package imports a concrete implementation — everything goes through `core` interfaces.
