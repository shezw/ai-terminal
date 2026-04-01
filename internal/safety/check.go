package safety

import (
	"path/filepath"
	"strings"

	"github.com/shezw/ai-terminal/internal/config"
)

// IsPathSafe checks whether a given absolute path is allowed.
// Denied list takes priority over allow list.
func IsPathSafe(absPath string, cfg *config.ExecSafety) bool {
	absPath = filepath.Clean(absPath)

	for _, denied := range cfg.DeniedList {
		d, err := config.AbsolutePath(denied)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, d) {
			return false
		}
	}

	for _, allow := range cfg.AllowList {
		a, err := config.AbsolutePath(allow)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, a) {
			return true
		}
	}

	return false
}
