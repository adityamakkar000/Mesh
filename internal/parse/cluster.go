package parse

import (
    "fmt"
    "net"
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
            var n_hosts = len(clusters[k].Hosts)
            if n_hosts == 0 {
                fmt.Printf("expected cluster %s to have at least 1 IP got 0\n", k)
                delete(clusters, k)
                continue
            }
            
            for _, host := range clusters[k].Hosts {
                if net.ParseIP(host) == nil {
                    fmt.Printf("invalid IP address %s in cluster %s\n", host, k)
                    delete(clusters, k)
                    break
                }
            }
            
            if _, exists := clusters[k]; exists {
                fmt.Printf("parsed cluster %s with %d hosts\n", k, n_hosts)
            }
        }


    return clusters, nil
}

