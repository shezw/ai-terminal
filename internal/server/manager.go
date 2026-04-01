package server

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/shezw/ai-terminal/internal/config"
	"github.com/shezw/ai-terminal/internal/render"
)

const pidFile = "llama-server.pid"

func pidFilePath() string {
	return config.AppDir() + "/" + pidFile
}

func Start(cfg *config.ServerConfig) error {
	if IsRunning() {
		render.PrintInfo("llama-server is already running.")
		return nil
	}

	llamaPath, err := exec.LookPath("llama-server")
	if err != nil {
		return fmt.Errorf("llama-server not found in PATH; install llama.cpp first")
	}

	if cfg.ModelPath == "" {
		return fmt.Errorf("no model configured; use 'ai-terminal model install' first")
	}

	args := []string{
		"-m", cfg.ModelPath,
		"--port", strconv.Itoa(cfg.Port),
		"-c", strconv.Itoa(cfg.ContextSize),
	}

	cmd := exec.Command(llamaPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start llama-server: %w", err)
	}

	if err := config.EnsureDir(config.AppDir()); err != nil {
		return err
	}
	pidData := []byte(strconv.Itoa(cmd.Process.Pid))
	if err := os.WriteFile(pidFilePath(), pidData, 0600); err != nil {
		return err
	}

	render.PrintInfo(fmt.Sprintf("llama-server started (PID %d) on port %d", cmd.Process.Pid, cfg.Port))
	return nil
}

func Stop() error {
	data, err := os.ReadFile(pidFilePath())
	if err != nil {
		return fmt.Errorf("no running server found")
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid PID file")
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		_ = proc.Kill()
	}

	_ = os.Remove(pidFilePath())
	render.PrintInfo("llama-server stopped.")
	return nil
}

func IsRunning() bool {
	data, err := os.ReadFile(pidFilePath())
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	return proc.Signal(syscall.Signal(0)) == nil
}
