package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func configDir() string {
	if runtime.GOOS == "windows" {
		dir, err := os.UserConfigDir()
		if err != nil {
			dir = filepath.Join(os.Getenv("APPDATA"), "mesh")
		}
		return filepath.Join(dir, "mesh")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic("cannot find home directory")
	}
	return filepath.Join(home, ".config", "mesh")
}

func cacheDir() string {
	return filepath.Join(configDir(), "cache")
}

func JobsDir() string {
	return filepath.Join(cacheDir(), "jobs")
}

func LogsDir(jobID string) string {
	return filepath.Join(JobsDir(), jobID, "logs")
}

func ConfigFile() string {
	return filepath.Join(configDir(), "mesh.yaml")
}
