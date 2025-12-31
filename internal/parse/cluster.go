package parse

import (
    "fmt"
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

type NodeConfig struct {
    User         string   `yaml:"user"`
    IdentityFile string   `yaml:"identity_file"`
    Hosts        []string `yaml:"hosts"`
}

type Clusters map[string]NodeConfig


func ParseClusters(path string) (Clusters, error) {
    filename, err := filepath.Abs(path)
    if err != nil {
        return nil, err
    }

    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var clusters Clusters
    if err := yaml.Unmarshal(data, &clusters); err != nil {
        return nil, err
    }

    for k := range clusters {
        if len(clusters[k].Hosts) == 0 {
            panic(fmt.Sprintf("expected cluster %s to have at least 1 IP", k))
        }
    }

    return clusters, nil
}

