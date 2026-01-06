package parse

import (
	"os"
	"path/filepath"

	"github.com/adityamakkar000/Mesh/internal/ui"
	"gopkg.in/yaml.v3"
)

type MeshConfig struct {
	Commands []string `yaml:"commands"`
	Ignore   []string `yaml:"ignore"`
	Prerun   []string `yaml:"prerun"`
}

func Mesh() (*MeshConfig, error) {
	filename, err := filepath.Abs("./mesh.yaml")
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var mesh MeshConfig
	if err := yaml.Unmarshal(data, &mesh); err != nil {
		return nil, err
	}

	if len(mesh.Commands) == 0 {
		ui.Warn("no 'commands' specified in mesh.yaml")
	}
	if len(mesh.Ignore) == 0 {
		ui.Warn("no 'ignore' patterns specified in mesh.yaml")
	}
	if len(mesh.Prerun) == 0 {
		ui.Warn("no 'prerun' commands specified in mesh.yaml")
	}

	return &mesh, nil
}
