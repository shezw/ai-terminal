package mode

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/shezw/ai-terminal/internal/config"
	"github.com/shezw/ai-terminal/internal/render"
	"github.com/shezw/ai-terminal/internal/server"
)

type ModelTier struct {
	Name     string
	Params   string
	Filename string
	URL      string
}

func RunModel(args []string) error {
	if len(args) == 0 {
		return modelMenu()
	}

	switch args[0] {
	case "config":
		return modelConfig()
	case "install":
		return modelInstall()
	case "list":
		return modelList()
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: ai-terminal model remove <model-name>")
		}
		return modelRemove(args[1])
	default:
		return fmt.Errorf("unknown model subcommand: %s", args[0])
	}
}

func modelMenu() error {
	fmt.Println("AI Terminal - Model Management")
	fmt.Println()
	fmt.Println("  config    - Configure API or local model settings")
	fmt.Println("  install   - Detect system and install a recommended model")
	fmt.Println("  list      - List installed models")
	fmt.Println("  remove    - Remove an installed model")
	fmt.Println()
	fmt.Println("Or configure model settings:")
	fmt.Println("  ai-terminal model config")
	fmt.Println("  Edit", config.ConfigFilePath())
	return nil
}

func modelConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := runModelConfigInteractive(os.Stdin, os.Stdout, cfg); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	render.PrintInfo("Config saved to: " + config.ConfigFilePath())
	return nil
}

func runModelConfigInteractive(in io.Reader, out io.Writer, cfg *config.Config) error {
	reader := bufio.NewReader(in)

	fmt.Fprintln(out, "AI Terminal - Model Configuration")
	fmt.Fprintf(out, "Config file: %s\n\n", config.ConfigFilePath())

	mode, err := promptChoice(reader, out, "Mode", cfg.Mode, []string{"api", "local"})
	if err != nil {
		return err
	}
	cfg.Mode = mode

	switch mode {
	case "api":
		endpoint, err := promptRequiredString(reader, out, "API endpoint", cfg.API.Endpoint)
		if err != nil {
			return err
		}
		apiKey, err := promptAPIKey(reader, out, cfg.API.APIKey)
		if err != nil {
			return err
		}
		model, err := promptRequiredString(reader, out, "Model name", cfg.API.Model)
		if err != nil {
			return err
		}

		cfg.API.Endpoint = endpoint
		cfg.API.APIKey = apiKey
		cfg.API.Model = model

		fmt.Fprintf(out, "\nRemote model configured: %s\n", cfg.API.Model)
	case "local":
		modelPath, err := promptRequiredString(reader, out, "Local model path", cfg.Server.ModelPath)
		if err != nil {
			return err
		}
		modelPath, err = config.AbsolutePath(modelPath)
		if err != nil {
			return err
		}

		if _, err := os.Stat(modelPath); err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(out, "Warning: model file not found yet: %s\n", modelPath)
			} else {
				return err
			}
		}

		port, err := promptPositiveInt(reader, out, "Server port", cfg.Server.Port)
		if err != nil {
			return err
		}
		contextSize, err := promptPositiveInt(reader, out, "Context size", cfg.Server.ContextSize)
		if err != nil {
			return err
		}

		cfg.Server.ModelPath = modelPath
		cfg.Server.Port = port
		cfg.Server.ContextSize = contextSize

		fmt.Fprintf(out, "\nLocal model configured: %s\n", cfg.Server.ModelPath)
	}

	return nil
}

func promptChoice(reader *bufio.Reader, out io.Writer, label, defaultValue string, allowed []string) (string, error) {
	for {
		value, err := promptString(reader, out, label, defaultValue)
		if err != nil {
			return "", err
		}
		value = strings.ToLower(strings.TrimSpace(value))
		for _, option := range allowed {
			if value == option {
				return value, nil
			}
		}
		fmt.Fprintf(out, "Invalid value %q. Allowed values: %s\n", value, strings.Join(allowed, "/"))
	}
}

func promptRequiredString(reader *bufio.Reader, out io.Writer, label, defaultValue string) (string, error) {
	for {
		value, err := promptString(reader, out, label, defaultValue)
		if err != nil {
			return "", err
		}
		if value != "" {
			return value, nil
		}
		fmt.Fprintf(out, "%s is required.\n", label)
	}
}

func promptPositiveInt(reader *bufio.Reader, out io.Writer, label string, defaultValue int) (int, error) {
	for {
		value, err := promptString(reader, out, label, strconv.Itoa(defaultValue))
		if err != nil {
			return 0, err
		}
		parsed, err := strconv.Atoi(value)
		if err == nil && parsed > 0 {
			return parsed, nil
		}
		fmt.Fprintf(out, "%s must be a positive integer.\n", label)
	}
}

func promptAPIKey(reader *bufio.Reader, out io.Writer, current string) (string, error) {
	prompt := "API key"
	if current != "" {
		prompt += " [press Enter to keep current]"
	} else {
		prompt += " [optional]"
	}
	prompt += ": "

	fmt.Fprint(out, prompt)
	value, err := readPromptLine(reader)
	if err != nil {
		return "", err
	}
	if value == "" && current != "" {
		return current, nil
	}
	return value, nil
}

func promptString(reader *bufio.Reader, out io.Writer, label, defaultValue string) (string, error) {
	prompt := label
	if defaultValue != "" {
		prompt += fmt.Sprintf(" [%s]", defaultValue)
	}
	prompt += ": "

	fmt.Fprint(out, prompt)
	value, err := readPromptLine(reader)
	if err != nil {
		return "", err
	}
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func readPromptLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF && line != "" {
			return strings.TrimSpace(line), nil
		}
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func detectSystemTier() string {
	arch := runtime.GOARCH
	goos := runtime.GOOS

	switch {
	case goos == "darwin" && arch == "arm64":
		return "medium"
	case goos == "linux":
		if _, err := exec.LookPath("nvidia-smi"); err == nil {
			return "high"
		}
		return "medium"
	default:
		return "low"
	}
}

const hfMirror = "https://huggingface.co"

func modelTiers() map[string]ModelTier {
	return map[string]ModelTier{
		"low": {
			Name:     "Qwen3-0.6B",
			Params:   "0.6B",
			Filename: "qwen3-0.6b-q4_k_m.gguf",
			URL:      hfMirror + "/Qwen/Qwen3-0.6B-GGUF/resolve/main/qwen3-0.6b-q4_k_m.gguf",
		},
		"medium": {
			Name:     "Qwen3-8B",
			Params:   "8B",
			Filename: "qwen3-8b-q4_k_m.gguf",
			URL:      hfMirror + "/Qwen/Qwen3-8B-GGUF/resolve/main/qwen3-8b-q4_k_m.gguf",
		},
		"high": {
			Name:     "Qwen3-32B",
			Params:   "32B",
			Filename: "qwen3-32b-q4_k_m.gguf",
			URL:      hfMirror + "/Qwen/Qwen3-32B-GGUF/resolve/main/qwen3-32b-q4_k_m.gguf",
		},
	}
}

func modelInstall() error {
	tier := detectSystemTier()
	tiers := modelTiers()
	recommended := tiers[tier]

	fmt.Printf("Detected system tier: %s\n", tier)
	fmt.Printf("Recommended model: %s (%s parameters)\n", recommended.Name, recommended.Params)
	fmt.Println()

	// Ensure llama-server
	llamaPath, err := server.EnsureLlamaServer()
	if err != nil {
		render.PrintWarning(fmt.Sprintf("Could not set up llama-server: %s", err))
		fmt.Println("You can also install llama.cpp manually: https://github.com/ggerganov/llama.cpp")
	} else {
		render.PrintInfo(fmt.Sprintf("llama-server ready at: %s", llamaPath))
	}
	fmt.Println()

	// Check if model already downloaded
	modelPath := filepath.Join(config.ModelsDir(), recommended.Filename)
	if _, err := os.Stat(modelPath); err == nil {
		render.PrintInfo(fmt.Sprintf("Model already exists: %s", modelPath))
		return updateConfigModel(modelPath)
	}

	fmt.Printf("Download %s (%s)? [Y/n] ", recommended.Name, recommended.Params)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input != "" && input != "y" && input != "yes" {
		fmt.Println("Skipped. You can place GGUF models manually in:", config.ModelsDir())
		return nil
	}

	// Download model
	if err := config.EnsureDir(config.ModelsDir()); err != nil {
		return err
	}

	render.PrintInfo(fmt.Sprintf("Downloading %s...", recommended.Name))
	if err := downloadModel(recommended.URL, modelPath); err != nil {
		return fmt.Errorf("download model: %w", err)
	}

	render.PrintInfo(fmt.Sprintf("Model saved to: %s", modelPath))
	return updateConfigModel(modelPath)
}

func updateConfigModel(modelPath string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Mode = "local"
	cfg.Server.ModelPath = modelPath
	if err := cfg.Save(); err != nil {
		return err
	}
	render.PrintInfo("Config updated: mode=local, model_path=" + modelPath)
	return nil
}

func downloadModel(url, dest string) error {
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
	buf := make([]byte, 64*1024)
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
				sizeMB := float64(written) / (1024 * 1024)
				totalMB := float64(size) / (1024 * 1024)
				fmt.Printf("\r  %.1f MB / %.1f MB (%.1f%%)", sizeMB, totalMB, pct)
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

func modelList() error {
	dir := config.ModelsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No models installed.")
			return nil
		}
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No models installed.")
		return nil
	}

	fmt.Println("Installed models:")
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".gguf") {
			info, _ := e.Info()
			size := ""
			if info != nil {
				sizeMB := info.Size() / (1024 * 1024)
				size = fmt.Sprintf(" (%d MB)", sizeMB)
			}
			fmt.Printf("  %s%s\n", e.Name(), size)
		}
	}
	return nil
}

func modelRemove(name string) error {
	path := filepath.Join(config.ModelsDir(), name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("model not found: %s", name)
	}
	return os.Remove(path)
}
