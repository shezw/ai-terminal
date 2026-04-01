package mode

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/shezw/ai-terminal/internal/config"
)

func LoadRem() (map[string]string, error) {
	path := config.LocalRemPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	kv := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			kv[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return kv, nil
}

func SaveRem(kv map[string]string) error {
	if err := config.EnsureDir(config.AppDir()); err != nil {
		return err
	}
	var sb strings.Builder
	for k, v := range kv {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	return os.WriteFile(config.LocalRemPath(), []byte(sb.String()), 0600)
}

func RunRem(args []string) error {
	if len(args) < 2 {
		kv, err := LoadRem()
		if err != nil {
			return err
		}
		if len(kv) == 0 {
			fmt.Println("No saved preferences.")
			return nil
		}
		for k, v := range kv {
			fmt.Printf("  %s = %s\n", k, v)
		}
		return nil
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	kv, err := LoadRem()
	if err != nil {
		return err
	}
	kv[key] = value
	if err := SaveRem(kv); err != nil {
		return err
	}

	fmt.Printf("Saved: %s = %s\n", key, value)
	return nil
}
