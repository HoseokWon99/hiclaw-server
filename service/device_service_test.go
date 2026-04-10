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
