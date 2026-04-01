package mode

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shezw/ai-terminal/internal/config"
	"github.com/shezw/ai-terminal/internal/render"
)

type ModelTier struct {
	Name   string
	Params string
}

func RunModel(args []string) error {
	if len(args) == 0 {
		return modelMenu()
	}

	switch args[0] {
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
	fmt.Println("  install   - Detect system and install a recommended model")
	fmt.Println("  list      - List installed models")
	fmt.Println("  remove    - Remove an installed model")
	fmt.Println()
	fmt.Println("Or configure an API endpoint:")
	fmt.Println("  Edit", config.ConfigFilePath())
	return nil
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

func modelTiers() map[string]ModelTier {
	return map[string]ModelTier{
		"low":    {Name: "Qwen2.5-0.5B", Params: "0.5B"},
		"medium": {Name: "Qwen2.5-7B", Params: "7B"},
		"high":   {Name: "Qwen2.5-32B", Params: "32B"},
	}
}

func modelInstall() error {
	tier := detectSystemTier()
	tiers := modelTiers()
	recommended := tiers[tier]

	fmt.Printf("Detected system tier: %s\n", tier)
	fmt.Printf("Recommended model: %s (%s parameters)\n", recommended.Name, recommended.Params)
	fmt.Println()

	if _, err := exec.LookPath("llama-server"); err != nil {
		render.PrintWarning("llama-server not found in PATH.")
		fmt.Println("Please install llama.cpp first: https://github.com/ggerganov/llama.cpp")
		return nil
	}

	render.PrintInfo("Model download and management is in development.")
	render.PrintInfo(fmt.Sprintf("For now, place GGUF models in: %s", config.ModelsDir()))

	return config.EnsureDir(config.ModelsDir())
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
	path := config.ModelsDir() + "/" + name
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("model not found: %s", name)
	}
	return os.Remove(path)
}
