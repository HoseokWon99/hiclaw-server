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
