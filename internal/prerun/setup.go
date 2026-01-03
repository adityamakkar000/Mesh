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

type SSHCommand func(ctx context.Context, cluster *parse.NodeConfig, mesh *parse.MeshConfig, host string) error

type PreRunSSHMsgs struct {
	HostSuccessMsg string
	HostErrorMsg   string
	SuccessMsg     string
	ErrorMsg       string
}

func RunPreRunSSH(clusterName string, fn SSHCommand, msgs PreRunSSHMsgs) error {
	clusters, err := parse.Clusters()
	if err != nil {
		return ui.ErrorWrap(err, "failed to parse cluster.yaml")
	}

	cluster, ok := clusters[clusterName]
	if !ok {
		return ui.ErrorWrap(fmt.Errorf("cluster not found"), "cluster '%s' not found in cluster.yaml", clusterName)
	}

	mesh, err := parse.Mesh()
	if err != nil {
		return ui.ErrorWrap(err, "failed to parse mesh.yaml")
	}

	if len(mesh.Commands) == 0 {
		ui.Info("No commands to run in mesh.yaml")
		return nil
	}

	ui.Header(fmt.Sprintf("Setting up cluster '%s' (%d hosts)", clusterName, len(cluster.Hosts)))

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

	for _, host := range cluster.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			if err := fn(ctx, &cluster, mesh, host); err != nil {
				mu.Lock()
				failures++
				mu.Unlock()
				// ui.Error(fmt.Sprintf("[%s] setup failed: %v", host, err))
				ui.Error(fmt.Sprintf(msgs.HostErrorMsg, host, err))
			} else {
				// ui.Success(fmt.Sprintf("[%s] setup completed", host))
				ui.Success(fmt.Sprintf(msgs.HostSuccessMsg, host))
			}
		}(host)
	}

	wg.Wait()
	signal.Stop(sigChan)

	if failures > 0 {
		return fmt.Errorf(msgs.ErrorMsg, failures)
	}

	ui.Success(fmt.Sprintf(msgs.SuccessMsg, clusterName))
	return nil
}
