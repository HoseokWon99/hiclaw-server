// Package core core/stores.go
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
