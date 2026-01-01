package cli

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/adityamakkar000/Mesh/internal/config"
	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/ssh"
	"github.com/adityamakkar000/Mesh/internal/ui"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup <cluster_name>",
	Short: "Setup a cluster from mesh.yaml configuration",
	Long: `Setup a cluster by reading cluster info from ~/.config/mesh/node.yaml
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
	clusters, err := parse.Clusters(config.ConfigFile())
	if err != nil {
		return ui.ErrorWrap(err, "failed to parse node.yaml")
	}

	cluster, ok := clusters[clusterName]
	if !ok {
		return ui.ErrorWrap(fmt.Errorf("cluster not found"), "cluster '%s' not found in node.yaml", clusterName)
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

	var wg sync.WaitGroup
	errChan := make(chan error, len(cluster.Hosts))

	for i, host := range cluster.Hosts {
		wg.Add(1)
		go func(idx int, host string) {
			defer wg.Done()

			var stdout, stderr io.Writer
			if idx == 0 {
				prefix := fmt.Sprintf("%s[%s]%s ", ui.Cyan, host, ui.Reset)
				stdout = ui.NewPrefixWriter(prefix, os.Stdout)
				stderr = ui.NewPrefixWriter(prefix, os.Stderr)
			} else {
				stdout = io.Discard
				stderr = io.Discard
			}

			if err := setupHost(cluster, mesh, host, stdout, stderr); err != nil {
				errChan <- fmt.Errorf("host %s: %w", host, err)
			}
		}(i, host)
	}

	wg.Wait()
	close(errChan)

	failures := 0
	for err := range errChan {
		failures++
		ui.Error(err.Error())
	}

	if failures > 0 {
		return fmt.Errorf("setup failed on %d host(s)", failures)
	}

	ui.Success(fmt.Sprintf("Cluster '%s' setup complete", clusterName))
	return nil
}

func setupHost(cluster parse.NodeConfig, mesh *parse.MeshConfig, host string, stdout, stderr io.Writer) error {
	client, err := ssh.Connect(cluster.User, host, cluster.IdentityFile)
	if err != nil {
		return ui.ErrorWrap(err, "failed to connect")
	}
	defer client.Close()

	for _, command := range mesh.Commands {
		if err := client.Exec(command, stdout, stderr); err != nil {
			return ui.ErrorWrap(err, "failed to execute '%s'", command)
		}
	}

	return nil
}
