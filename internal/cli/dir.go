package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// meshYAMLDirFlag is set by --dir on run and setup: directory that contains mesh.yaml
// (empty means current directory).
var meshYAMLDirFlag string

func resolveMeshYAMLDir() (string, error) {
	if meshYAMLDirFlag == "" {
		return filepath.Abs(".")
	}
	abs, err := filepath.Abs(meshYAMLDirFlag)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("mesh.yaml directory: %w", err)
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("mesh.yaml path is not a directory: %s", abs)
	}
	return abs, nil
}
