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
