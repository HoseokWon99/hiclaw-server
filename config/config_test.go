package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := Load()
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("expected :8080, got %s", cfg.HTTPAddr)
	}
	if cfg.SQLitePath != "devices.db" {
		t.Errorf("expected devices.db, got %s", cfg.SQLitePath)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("HICLAW_HTTP_ADDR", ":9090")
	defer os.Unsetenv("HICLAW_HTTP_ADDR")

	cfg := Load()
	if cfg.HTTPAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.HTTPAddr)
	}
}
