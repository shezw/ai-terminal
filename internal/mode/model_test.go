package mode

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shezw/ai-terminal/internal/config"
)

func TestRunModelConfigInteractiveAPI(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.APIKey = "sk-current"

	input := strings.NewReader("\nhttps://example.com/v1\n\nqwen-plus\n")
	var output bytes.Buffer

	if err := runModelConfigInteractive(input, &output, cfg); err != nil {
		t.Fatalf("runModelConfigInteractive() error = %v", err)
	}

	if cfg.Mode != "api" {
		t.Fatalf("expected mode api, got %q", cfg.Mode)
	}
	if cfg.API.Endpoint != "https://example.com/v1" {
		t.Fatalf("expected endpoint updated, got %q", cfg.API.Endpoint)
	}
	if cfg.API.APIKey != "sk-current" {
		t.Fatalf("expected api key to be preserved, got %q", cfg.API.APIKey)
	}
	if cfg.API.Model != "qwen-plus" {
		t.Fatalf("expected model updated, got %q", cfg.API.Model)
	}
	if !strings.Contains(output.String(), "AI Terminal - Model Configuration") {
		t.Fatalf("expected interactive header in output, got %q", output.String())
	}
	if strings.Contains(output.String(), "sk-current") {
		t.Fatalf("expected api key to remain hidden in output, got %q", output.String())
	}
}

func TestRunModelConfigInteractiveLocal(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	modelFile := filepath.Join(tempDir, "local-model.gguf")
	if err := os.WriteFile(modelFile, []byte("stub"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := config.DefaultConfig()
	input := strings.NewReader("invalid\nlocal\nlocal-model.gguf\n9001\n4096\n")
	var output bytes.Buffer

	if err := runModelConfigInteractive(input, &output, cfg); err != nil {
		t.Fatalf("runModelConfigInteractive() error = %v", err)
	}

	if cfg.Mode != "local" {
		t.Fatalf("expected mode local, got %q", cfg.Mode)
	}
	if cfg.Server.ModelPath != modelFile {
		t.Fatalf("expected absolute model path %q, got %q", modelFile, cfg.Server.ModelPath)
	}
	if cfg.Server.Port != 9001 {
		t.Fatalf("expected port 9001, got %d", cfg.Server.Port)
	}
	if cfg.Server.ContextSize != 4096 {
		t.Fatalf("expected context size 4096, got %d", cfg.Server.ContextSize)
	}
	if !strings.Contains(output.String(), "Invalid value \"invalid\"") {
		t.Fatalf("expected invalid mode warning, got %q", output.String())
	}
}
