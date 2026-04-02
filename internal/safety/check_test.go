package safety

import (
	"testing"

	"github.com/shezw/ai-terminal/internal/config"
)

func TestIsPathSafe_AllowedPath(t *testing.T) {
	cfg := &config.ExecSafety{
		AllowList:  []string{"/home/user"},
		DeniedList: []string{},
	}
	if !IsPathSafe("/home/user/documents/file.txt", cfg) {
		t.Error("expected path under allow list to be safe")
	}
}

func TestIsPathSafe_DeniedPath(t *testing.T) {
	cfg := &config.ExecSafety{
		AllowList:  []string{"/home/user"},
		DeniedList: []string{"/home/user/secret"},
	}
	if IsPathSafe("/home/user/secret/key.pem", cfg) {
		t.Error("expected path under denied list to be unsafe")
	}
}

func TestIsPathSafe_DeniedPriority(t *testing.T) {
	cfg := &config.ExecSafety{
		AllowList:  []string{"/home/user"},
		DeniedList: []string{"/home/user"},
	}
	if IsPathSafe("/home/user/file.txt", cfg) {
		t.Error("denied list should take priority over allow list")
	}
}

func TestIsPathSafe_UnlistedPath(t *testing.T) {
	cfg := &config.ExecSafety{
		AllowList:  []string{"/home/user"},
		DeniedList: []string{},
	}
	if IsPathSafe("/etc/passwd", cfg) {
		t.Error("path not in allow list should be unsafe")
	}
}

func TestIsPathSafe_EmptyLists(t *testing.T) {
	cfg := &config.ExecSafety{
		AllowList:  []string{},
		DeniedList: []string{},
	}
	if IsPathSafe("/any/path", cfg) {
		t.Error("empty allow list should deny all paths")
	}
}
