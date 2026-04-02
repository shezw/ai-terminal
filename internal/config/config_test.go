package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Mode != "api" {
		t.Errorf("expected default mode api, got %q", cfg.Mode)
	}
	if cfg.Server.Port != 8398 {
		t.Errorf("expected default port 8398, got %d", cfg.Server.Port)
	}
	if cfg.Language != "en" {
		t.Errorf("expected default language en, got %q", cfg.Language)
	}
	if len(cfg.Safety.AllowList) == 0 {
		t.Error("expected default allow list to contain home dir")
	}
}

func TestAbsolutePath_Absolute(t *testing.T) {
	p, err := AbsolutePath("/usr/local/bin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "/usr/local/bin" {
		t.Errorf("expected /usr/local/bin, got %q", p)
	}
}

func TestAbsolutePath_Relative(t *testing.T) {
	p, err := AbsolutePath("relative/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(p) {
		t.Errorf("expected absolute path, got %q", p)
	}
}

func TestEnsureDir(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "ai-terminal-test-dir", "sub")
	defer os.RemoveAll(filepath.Join(os.TempDir(), "ai-terminal-test-dir"))
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
}

func TestHomeDir(t *testing.T) {
	home := HomeDir()
	if home == "" {
		t.Error("expected non-empty home directory")
	}
}
