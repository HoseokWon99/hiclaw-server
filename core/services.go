// Package core core/services.go
package core

type ChatService interface {
	SendMessage(sender *Device, content []*Part) (*Message, error)
	HandleAgentReply(content []*Part) (*Message, error)
	ListRecent(limit int) ([]*Message, error)
}

type AgentService interface {
	Notify(msg *Message) error
}

type DeviceService interface {
	Register(device *Device) error
	Remove(ip string) error
	ListAll() ([]*Device, error)
	GetAgent() (*Device, error)
}
