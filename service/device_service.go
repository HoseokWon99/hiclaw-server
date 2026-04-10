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
