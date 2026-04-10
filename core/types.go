package core

import (
	"net"
	"time"
)

type Device struct {
	IP   *net.IPAddr `json:"ip"`
	Name string      `json:"name"`
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
