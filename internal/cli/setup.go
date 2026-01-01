package cli

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/ssh"
	"github.com/adityamakkar000/Mesh/internal/ui"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup <cluster_path> <cluster_name>",
	Short: "Setup a cluster from a YAML configuration",
	Long: `Setup a cluster by parsing the YAML configuration file and executing setup commands.

Example:
  mesh setup ./cluster.yaml my-cluster`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clusterPath := args[0]
		clusterName := args[1]

		if err := runSetup(clusterPath, clusterName); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(clusterPath, clusterName string) error {
	clusters, err := parse.Clusters(clusterPath)
	if err != nil {
		return ui.ErrorWrap(err, "failed to parse clusters")
	}

	cluster, ok := clusters[clusterName]
	if !ok {
		return ui.ErrorWrap(err, "cluster '%s' not found in %s", clusterName, clusterPath)
	}

	if len(cluster.Commands) == 0 {
		ui.Info(fmt.Sprintf("No commands to run for cluster '%s'", clusterName))
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

			if err := setupHost(cluster, host, stdout, stderr); err != nil {
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

func setupHost(cluster parse.NodeConfig, host string, stdout, stderr io.Writer) error {
	client, err := ssh.Connect(cluster.User, host, cluster.IdentityFile)
	if err != nil {
		return ui.ErrorWrap(err, "failed to connect")
	}
	defer client.Close()

	for _, command := range cluster.Commands {
		if err := client.Exec(command, stdout, stderr); err != nil {
			return ui.ErrorWrap(err, "failed to execute '%s'", command)
		}
	}

	return nil
}
