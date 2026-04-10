package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hiclaw-server/core"
)

type AgentServiceImpl struct {
	webhookURL string
	http       *http.Client
}

func NewAgentService(baseURL, webhookPath string) *AgentServiceImpl {
	return &AgentServiceImpl{
		webhookURL: baseURL + webhookPath,
		http:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *AgentServiceImpl) Notify(msg *core.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	res, err := s.http.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send to agent: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("agent returned %d", res.StatusCode)
	}
	return nil
}
