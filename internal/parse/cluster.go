package parse

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/adityamakkar000/Mesh/internal/config"
	"github.com/adityamakkar000/Mesh/internal/ui"
	"gopkg.in/yaml.v3"
)

type NodeConfig struct {
	User         string   `yaml:"user"`
	IdentityFile string   `yaml:"identity_file"`
	Hosts        []string `yaml:"hosts"`
}

type ClusterMap map[string]NodeConfig

func Clusters() (ClusterMap, error) {
	filename, err := filepath.Abs(config.ClusterFile())
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var clusters ClusterMap
	if err := yaml.Unmarshal(data, &clusters); err != nil {
		return nil, err
	}

	for k := range clusters {
		var n_hosts = len(clusters[k].Hosts)
		if n_hosts == 0 {
			ui.Error(fmt.Sprintf("expected cluster %s to have at least 1 IP got 0", k))
			delete(clusters, k)
			continue
		}

		for _, host := range clusters[k].Hosts {
			if net.ParseIP(host) == nil {
				ui.Error(fmt.Sprintf("invalid IP address %s in cluster %s", host, k))
				delete(clusters, k)
				break
			}
		}

		if _, exists := clusters[k]; exists {
			ui.Info(fmt.Sprintf("parsed cluster %s with %d hosts", k, n_hosts))
		}
	}

	var cluster_names []string
	for k := range clusters {
		cluster_names = append(cluster_names, k)
	}
	ui.Info(fmt.Sprintf("available clusters: %v", cluster_names))

	return clusters, nil
}
