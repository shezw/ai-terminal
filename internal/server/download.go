package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/shezw/ai-terminal/internal/config"
	"github.com/shezw/ai-terminal/internal/render"
)

// LlamaServerBinPath returns the path to the managed llama-server binary.
func LlamaServerBinPath() string {
	return config.AppDir() + "/bin/llama-server"
}

// LlamaServerRelease returns the download URL for the current platform.
func LlamaServerRelease() (string, error) {
	base := "https://github.com/ggerganov/llama.cpp/releases/latest/download"
	goos := runtime.GOOS
	arch := runtime.GOARCH

	switch {
	case goos == "darwin" && arch == "arm64":
		return base + "/llama-server-macos-arm64", nil
	case goos == "darwin" && arch == "amd64":
		return base + "/llama-server-macos-x86_64", nil
	case goos == "linux" && arch == "amd64":
		return base + "/llama-server-linux-x86_64", nil
	case goos == "linux" && arch == "arm64":
		return base + "/llama-server-linux-aarch64", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, arch)
	}
}

// EnsureLlamaServer checks if llama-server exists, downloads if not.
func EnsureLlamaServer() (string, error) {
	binPath := LlamaServerBinPath()
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	render.PrintInfo("llama-server not found. Downloading...")

	url, err := LlamaServerRelease()
	if err != nil {
		return "", err
	}

	if err := config.EnsureDir(config.AppDir() + "/bin"); err != nil {
		return "", err
	}

	if err := downloadFile(url, binPath); err != nil {
		return "", fmt.Errorf("download llama-server: %w", err)
	}

	if err := os.Chmod(binPath, 0755); err != nil {
		return "", err
	}

	render.PrintInfo("llama-server downloaded successfully.")
	return binPath, nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	size := resp.ContentLength
	var written int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			nw, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			written += int64(nw)
			if size > 0 {
				pct := float64(written) / float64(size) * 100
				fmt.Printf("\r  Downloading... %.1f%%", pct)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}
	fmt.Println()
	return nil
}
