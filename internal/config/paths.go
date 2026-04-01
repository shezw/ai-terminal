package config

import (
	"os"
	"path/filepath"
)

const AppDirName = ".ai-terminal"

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func AppDir() string {
	return filepath.Join(HomeDir(), AppDirName)
}

func ConfigFilePath() string {
	return filepath.Join(AppDir(), "config.yaml")
}

func LocalRemPath() string {
	return filepath.Join(AppDir(), "local-rem")
}

func ModelsDir() string {
	return filepath.Join(AppDir(), "models")
}

func TemplatesDir() string {
	return filepath.Join(AppDir(), "templates")
}

// AbsolutePath converts a potentially relative path to absolute based on cwd.
func AbsolutePath(p string) (string, error) {
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(cwd, p)), nil
}

// EnsureDir creates a directory and all parents if they don't exist.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
