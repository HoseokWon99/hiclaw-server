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
