package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/ssh"
	"github.com/adityamakkar000/Mesh/internal/ui"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup <cluster_name>",
	Short: "Setup a cluster from mesh.yaml configuration",
	Long: `Setup a cluster by reading cluster info from ~/.config/mesh/cluster.yaml
and setup commands from ./mesh.yaml.

Example:
  mesh setup my-cluster`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]

		if err := runSetup(clusterName); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(clusterName string) error {
	clusters, err := parse.Clusters()
	if err != nil {
		return ui.ErrorWrap(err, "failed to parse cluster.yaml")
	}

	cluster, ok := clusters[clusterName]
	if !ok {
		return ui.ErrorWrap(fmt.Errorf("cluster not found"), "cluster '%s' not found in cluster.yaml", clusterName)
	}

	mesh, err := parse.Mesh("mesh.yaml")
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

			if err := setupHost(ctx, &cluster, mesh, host); err != nil {
				mu.Lock()
				failures++
				mu.Unlock()
				ui.Error(fmt.Sprintf("[%s] failed: %v", host, err))
			} else {
				ui.Success(fmt.Sprintf("[%s] setup completed", host))
			}
		}(host)
	}

	wg.Wait()
	signal.Stop(sigChan)

	if failures > 0 {
		return fmt.Errorf("setup failed on %d host(s)", failures)
	}

	ui.Success(fmt.Sprintf("Cluster '%s' setup complete", clusterName))
	return nil
}

func setupHost(ctx context.Context, cluster *parse.NodeConfig, mesh *parse.MeshConfig, host string) error {
	client, err := ssh.Connect(ctx, cluster.User, host, cluster.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	for _, command := range mesh.Commands {
		if err := client.Exec(ctx, command, io.Discard, io.Discard); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", command, err)
		}
	}

	return nil
}
