package prerun

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/ui"
)

type SSHCommand func(ctx context.Context, cluster *parse.NodeConfig, mesh *parse.MeshConfig, host string, host_id int) error

func ParseConfigs(clusterName string) (*parse.NodeConfig, *parse.MeshConfig, error) {
	clusters, err := parse.Clusters()
	if err != nil {
		return nil, nil, ui.ErrorWrap(err, "failed to parse cluster.yaml")
	}

	cluster, ok := (clusters)[clusterName]
	if !ok {
		return nil, nil, ui.ErrorWrap(fmt.Errorf("cluster not found"), "cluster '%s' not found in cluster.yaml", clusterName)
	}

	mesh, err := parse.Mesh()
	if err != nil {
		return nil, nil, ui.ErrorWrap(err, "failed to parse mesh.yaml")
	}

	if len(mesh.Commands) == 0 {
		ui.Info("No commands to run in mesh.yaml")
	}

	return &cluster, mesh, nil

}

func RunOnAllHosts(cluster *parse.NodeConfig, mesh *parse.MeshConfig, mainfn SSHCommand, sucess_message, error_message string) int {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChan:
			ui.Warn("Interrupt received, canceling all connections...")
			cancel()
		case <-ctx.Done():
		}
	}()

	var wg sync.WaitGroup
	var mu sync.Mutex
	failures := 0

	for id, host := range cluster.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			if err := mainfn(ctx, cluster, mesh, host, id); err != nil {
				mu.Lock()
				failures++
				mu.Unlock()
				ui.Error(fmt.Sprintf(error_message, host, err))
			} else {
				ui.Success(fmt.Sprintf(sucess_message, host))
			}
		}(host)
	}

	wg.Wait()
	signal.Stop(sigChan)

	return failures
}
