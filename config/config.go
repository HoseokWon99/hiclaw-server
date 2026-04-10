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
